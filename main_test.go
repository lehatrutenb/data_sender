package data_sender

import (
	client "data_sender/client"
	server "data_sender/server"
	tools "data_sender/tools"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

const address = "127.0.0.1:8182"

//tools.Cmp_slices(parsed_data, data_part)

func Cmp_slices[T tools.Number](a, b []T) bool { // true if equal, otherwise false
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func gen_byte_arr(len int64) []byte {
	arr := make([]byte, len)
	for i := int64(0); i < len; i++ {
		arr[i] = byte(rand.Intn(256))
	}
	return arr
}

func gen_data() [][][]byte {
	data := make([][][]byte, 1e4)
	for i := 0; i < len(data); i++ {
		if i < 10 {
			data[i] = [][]byte{gen_byte_arr(1e5)}
		} else if i < 1000 {
			data[i] = [][]byte{gen_byte_arr(1e4)}
		} else {
			data[i] = [][]byte{gen_byte_arr(1e2)}
		}
	}
	return data
}

const MaxTime = 1

func TestSend_tcp(t *testing.T) {
	test_data := gen_data()
	for k, data := range test_data {
		fmt.Println(k)
		queue := make(chan byte, 1)
		tcp_listener_ready := make(chan byte, len(data))

		timer := time.After(MaxTime * time.Second)
		go server.Get_tcps(address, int64(len(data)), false, queue, &tcp_listener_ready)
		ind := 0
		for _, data_part := range data {
			data_now := make([]byte, len(data_part))
			copy(data_now, data_part)
			<-tcp_listener_ready
			go client.Send_tcp(&data_now, address, tools.Instruction{1, "main_test/data" + fmt.Sprint(ind), int64(len(data_part)), 0, []byte{}, 0}, 1000)
			ind++
		}

		select {
		case <-timer:
			t.Error("\nTime limit exceeded\n")
			return
		case <-queue:
		}

		//intln("data len:", len(data))
		ind = 0
		for _, data_part := range data {
			//fmt.Println(ind)
			parsed_data, err := os.ReadFile("main_test/data" + fmt.Sprint(ind))
			ind++
			if err != nil {
				t.Error(
					"\nDuring reading file error occured:\n",
					err,
				)
				return
			}
			if !Cmp_slices(parsed_data, data_part) {
				t.Error(
					"\nArrays are different\n",
				)
				return
			}
		}
	}
}
