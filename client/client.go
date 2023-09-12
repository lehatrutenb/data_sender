package client

import (
	tools "data_sender/tools"
	"log"
	"math/rand"
	"net"
	"time"
)

func write_to_conn(data []byte, conn *net.UDPConn, queue chan byte) error {
	defer tools.Write_chan(queue)
	_, err := conn.Write(data)
	if err != nil {
		log.Print(err)
		return err
	}

	return err
}

const request_size = 8970
const sign_size = 6  // 255^6
const index_size = 4 // 255^4

func send_udp(data *[]byte, address string, inst tools.Instruction) error {
	s, err := net.ResolveUDPAddr("udp4", address)
	conn, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		log.Print(err)
		return err
	}
	//defer conn.Close() some goroutines can't send because connections closes too fast

	sign := make([]byte, sign_size)
	for i := 0; i < sign_size; i++ {
		sign[i] = byte(rand.Int() % 256)
	}

	kol_reqs := int64((len(*data) + request_size - 1) / request_size)

	num_reqs, err := tools.Int64_to_byte_arr_fixed_size(tools.Max_reqs_size, kol_reqs)
	if err != nil {
		log.Print(err)
		return err
	}

	queue := make(chan byte, kol_reqs)
	sk := tools.Skin_cl{sign, num_reqs, []byte{}}

	byte_inst, err := tools.Parse_instruction(inst)
	if err != nil {
		log.Print(err)
		return err
	}

	tools.Add_skin(&byte_inst, &sk)
	go write_to_conn(append(byte_inst, 255), conn, queue) // send instructions
	sk.Index = make([]byte, index_size)
	for i := 0; i < len(*data); i += request_size {
		buffer := (*data)[i:tools.Min(len(*data), i+request_size)]
		tools.Add_skin(&buffer, &sk)
		go write_to_conn(append(buffer, 0), conn, queue)

		for j := 0; j < len(sk.Index); j++ { // index in 0 it
			if sk.Index[j] == 255 {
				sk.Index[j] = 0
			} else {
				sk.Index[j]++
				break
			}
		}
	}

	for i := int64(0); i < kol_reqs; i++ {
		<-queue
	}

	return nil
}

func Send_tcp(data *[]byte, address string, inst tools.Instruction, timeout int) error {
	var conn net.Conn
	var err error
	conn, err = net.DialTimeout("tcp", address, time.Duration(timeout)*time.Millisecond)
	if err != nil {
		log.Print(err)
		return err
	}
	inst_byte, err := tools.Parse_instruction(inst)
	if err != nil {
		log.Print(err)
		return err
	}
	conn.Write(inst_byte)
	conn.Write(*data)
	conn.Close()
	return nil
}

/*func Send_tcp_mp(data *[]byte, address string, inst tools.Instruction, process_amount int) error {
	inst.Sign = make([]byte, sign_size)
	for i := 0; i < sign_size; i++ {
		inst.Sign[i] = byte(rand.Int() % 256)
	}

	inst.Req_amount = int64((len(*data) + request_size - 1) / request_size)
	for i := int64(0); i < inst.Req_amount; i++ {
		inst.Data_ind = i
		go Send_tcp(data, address, inst)
	}
	return nil
}*/

func main() {
	data := make([]byte, 1e9)
	for i := 0; i < len(data); i++ {
		data[i] = byte(i % 256)
	}
	//data, _ := os.ReadFile("video.mp4")
	//data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	/*if err != nil {
		log.Print(err)
		return
	}*/
	//send_udp(&data, "127.0.0.1:8181", tools.Instruction{1, "data2", -1, -1, []byte{}, -1})
	//send_tcp(&data, "127.0.0.1:8181", instruction{1, "data1", int64(len(data)), -1, []byte{}, -1})
	//send_tcp_mp(&data, "127.0.0.1:8181", instruction{1, "data1", int64(len(data)), -1, []byte{}, -1}, 1e6)
}
