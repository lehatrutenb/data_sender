package server

import (
	"bytes"
	tools "data_sender/tools"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

const max_sign_size = 6
const min_sign_size = 2
const udp_buffer_size = 9000

const will_use_udp = false

var sign_to_ind map[[max_sign_size]byte]int64
var free_ind = int64(0)
var global_data [][][]byte
var global_data_mex []int64

var global_data_instructions []tools.Instruction

func parse_request(buffer *[]byte, left, n int) (data []byte) {
	return (*buffer)[left:n]
}

func parse_skin(buffer *[]byte) (sk *tools.Skin_ser, left int) {
	now := 0

	sk = &tools.Skin_ser{}
	sk.Sign = (*buffer)[now+1 : now+int((*buffer)[now])+1]
	now += int((*buffer)[now]) + 1
	sk.Req_amount = tools.Byte_array_to_int64((*buffer)[now+1 : now+int((*buffer)[now])+1])
	now += int((*buffer)[now]) + 1
	sk.Index = tools.Byte_array_to_int64((*buffer)[now+1 : now+int((*buffer)[now])+1])
	now += int((*buffer)[now]) + 1
	return sk, now
}

type handle_data_int interface {
}

var Invalid_path = errors.New("Invalid path")

func create_path_to_file(path *string) error {
	if *path == "" {
		return Invalid_path
	}
	if (*path)[len(*path)-1] == '/' {
		*path = (*path)[:len(*path)-1]
	}

	filename_len := 0
	for i := len(*path) - 1; i >= 0; i-- {
		if (*path)[i] == '/' {
			break
		}
		filename_len++
	}
	if filename_len == 0 {
		return Invalid_path
	}
	err := os.Mkdir((*path)[:len(*path)-filename_len], 0750)
	if err != nil { // is it ok to change errors like there?
		return Invalid_path
	}
	return nil
}

func handle_data(inst tools.Instruction, data *handle_data_int) error {
	switch inst.Todo {
	// some data come, ready to save
	case 1:
		create_path_to_file(&inst.Name)
		file, err := os.Create(inst.Name)
		if err != nil {
			log.Print(err)
			file.Close()
			return err
		}
		switch (*data).(type) {
		case []byte:
			file.Write((*data).([]byte))
		case [][]byte:
			file.Write(bytes.Join((*data).([][]byte), nil))
		}
		file.Close()
	}
	return nil
}

func work_with_udp_req(buffer []byte, n int, queue *chan byte, mu *sync.Mutex) error {
	defer tools.Write_chan(*queue)

	var inst tools.Instruction
	var data []byte
	is_instruction := (buffer[n-1] == 255)
	n--

	sk, left := parse_skin(&buffer)
	if is_instruction {
		inst, _ = tools.Parse_instruction_request(&buffer, left)
	} else {
		data = parse_request(&buffer, left, n)
		for len(sk.Sign) < max_sign_size {
			sk.Sign = append(sk.Sign, 0)
		}
	}
	sign_fixed := [max_sign_size]byte(sk.Sign)

	if _, err := sign_to_ind[sign_fixed]; !err { // if there is no such sign
		if free_ind == int64(len(global_data)) {
			return errors.New("global_data is fullfilled")
		}
		global_data[free_ind] = make([][]byte, sk.Req_amount+1)
		sign_to_ind[sign_fixed] = free_ind
		free_ind++
	}

	global_data_index := sign_to_ind[sign_fixed]
	if is_instruction {
		global_data_instructions[global_data_index] = inst
	} else {
		mu.Lock()
		global_data[global_data_index][sk.Index] = data
		for len(global_data[global_data_index][global_data_mex[global_data_index]]) != 0 { // want to move mex
			global_data_mex[global_data_index]++
		}
		mu.Unlock()
	}

	mu.Lock()
	if global_data_mex[global_data_index] == int64(len(global_data[global_data_index]))-1 { // !!!!!!!!! CAN BE BAD EVERYWHERE IF LEN > 1E9
		global_data_mex[global_data_index]++
		data_to_handle := handle_data_int(global_data[global_data_index])
		err := handle_data(global_data_instructions[global_data_index], &data_to_handle)
		if err != nil {
			log.Print(err)
			return err
		}
	}
	mu.Unlock()

	return nil
}

func get_udps(address string, amount int) error { // amount = -1 - infinitive
	s, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		log.Print(err)
		return err
	}

	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		log.Print(err)
		return err
	}

	defer connection.Close()

	var mu sync.Mutex
	if amount == -1 {
		for {
			buffer := make([]byte, udp_buffer_size)
			n, _, err := connection.ReadFromUDP(buffer)

			if err != nil {
				log.Print(err)
				return err
			}
			go work_with_udp_req(buffer, n, nil, &mu)
		}
	} else {
		queue := make(chan byte, amount)
		for i := 0; i < amount; i++ {
			buffer := make([]byte, udp_buffer_size)
			n, _, err := connection.ReadFromUDP(buffer)

			if err != nil {
				log.Print(err)
				return err
			}
			go work_with_udp_req(buffer, n, &queue, &mu)
		}
		for i := 0; i < amount; i++ {
			_ = <-queue
		}
	}
	return nil
}

const tcp_buffer_size = 1e5 + 5e4

func handle_tcp_Connection(conn net.Conn, queue *chan byte) error {
	defer tools.Write_chan(*queue)
	defer conn.Close()
	var inst tools.Instruction
	var data []byte
	buffer := make([]byte, tcp_buffer_size)
	for i := 0; ; i++ {
		n, err := conn.Read(buffer)
		if err == io.EOF || n == 0 {
			data_to_handle := handle_data_int(data)
			handle_data(inst, &data_to_handle)
			return nil
		}
		if err != nil {
			log.Print(err)
			return err
		}

		if i == 0 {
			var left int
			inst, left = tools.Parse_instruction_request(&buffer, 0)
			data = make([]byte, 0, inst.Data_size)
			if left != n {
				data = append(data, parse_request(&buffer, left, n)...)
			}
		} else {
			data = append(data, parse_request(&buffer, 0, n)...)
		}
	}
	return nil
}

//const will_use_mp_tcp = true

const mp_tcp_buffer_size = 1e3

var global_data_chan []*chan byte

func handle_tcp_Connection_mp(conn net.Conn, queue *chan byte, mu *sync.Mutex) error {
	defer tools.Write_chan(*queue)
	defer conn.Close()
	var inst tools.Instruction
	buffer := make([]byte, tcp_buffer_size)
	var global_data_index int64
	var request_queue *chan byte

	for i := 0; ; i++ {
		n, err := conn.Read(buffer)
		if err == io.EOF || n == 0 {
			return nil
		}
		if err != nil {
			log.Print(err)
			return err
		}

		if i == 0 {
			inst, _ = tools.Parse_instruction_request(&buffer, 0) // may be there neeed some fix !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

			sign_fixed := [max_sign_size]byte(inst.Sign)
			if _, err := sign_to_ind[sign_fixed]; !err { // if there is no such sign
				if free_ind == int64(len(global_data)) {
					return errors.New("global_data is fullfilled")
				}
				global_data[free_ind] = make([][]byte, inst.Req_amount)
				sign_to_ind[sign_fixed] = free_ind
				free_ind++
			}

			global_data_index = int64(sign_to_ind[sign_fixed])
			if len(global_data[global_data_index]) == 0 {
				global_data[global_data_index] = make([][]byte, inst.Req_amount)
			}
			global_data[global_data_index][inst.Data_ind] = make([]byte, 0, inst.Data_size)

			mu.Lock()
			request_queue = global_data_chan[global_data_index]
			mu.Unlock()
			defer tools.Write_chan(*request_queue)
		} else {
			global_data[global_data_index][inst.Data_ind] = append(global_data[global_data_index][inst.Data_ind], parse_request(&buffer, 0, n)...)

			if inst.Data_ind == inst.Data_size-1 {
				tools.Write_chan(*request_queue)
				for i := int64(0); i < inst.Data_size; i++ {
					<-*request_queue
				}
			} else {
				continue
			}

			data_to_handle := handle_data_int(global_data[global_data_index])
			err = handle_data(global_data_instructions[global_data_index], &data_to_handle)
			if err != nil {
				log.Print(err)
				return err
			}
		}
	}
	return nil
}

func create_tcp_worker(ln *net.Listener, queue *chan byte, mu *sync.Mutex, will_use_mp_tcp bool, tcp_listener_ready *chan byte, address string) error {
	tools.Write_chan(*tcp_listener_ready)
	conn, err := (*ln).Accept()
	if err != nil {
		log.Print(err)
		return err
	}
	if will_use_mp_tcp {
		err = handle_tcp_Connection_mp(conn, queue, mu)
	} else {
		err = handle_tcp_Connection(conn, queue)
	}
	return err
}

func get_tcps(address string, amount int64, will_use_mp_tcp bool, tcp_listener_ready *chan byte) error { // amount = -1 - infinitive
	ln, err := net.Listen("tcp", address)

	if err != nil {
		log.Print(err)
		return err
	}
	defer ln.Close()

	var mu sync.Mutex
	if amount == -1 {
		for {
			go create_tcp_worker(&ln, nil, &mu, will_use_mp_tcp, tcp_listener_ready, address)
			if err != nil {
				log.Print(err)
				return err
			}
		}
	} else {
		queue := make(chan byte, amount)
		for i := int64(0); i < amount; i++ {
			go create_tcp_worker(&ln, &queue, &mu, will_use_mp_tcp, tcp_listener_ready, address)
			if err != nil {
				log.Print(err)
				return err
			}
		}
		for i := int64(0); i < amount; i++ {
			<-queue
		}
	}
	return nil
}

func Get_tcps(address string, amount int64, will_use_mp_tcp bool, ch chan byte, tcp_listener_ready *chan byte) error {
	global_data = make([][][]byte, amount)
	if will_use_mp_tcp {
		global_data_chan = make([]*chan byte, amount)
	}
	defer tools.Write_chan(ch)
	return get_tcps(address, amount, will_use_mp_tcp, tcp_listener_ready)
}

/*func main() {
	if will_use_udp || will_use_mp_tcp {
		global_data = make([][][]byte, 1e7)
		if will_use_udp {
			global_data_instructions = make([]tools.Instruction, 1e7)
			global_data_mex = make([]int64, 1e7)
		} else {
			global_data_chan = make([]*chan byte, 1e7)
		}
	}

	get_udps("127.0.0.1:8181", -1)
	//get_tcps("127.0.0.1:8181", -1)
}*/
