package tools

import (
	"fmt"
	"math/rand"
	"testing"
)

func gen_int64_signed() int64 {
	x := rand.Int63()
	if rand.Intn(2) == 0 {
		x *= -1
	}
	return x
}

func gen_byte_arr(amount int) []byte {
	res := make([]byte, amount)
	for i := 0; i < amount; i++ {
		res[i] = byte(rand.Intn(256))
	}
	return res
}

func gen_maxmin_tests() [][]int64 {
	res := make([][]int64, 1e5)
	for i := 0; i < len(res); i++ {
		res[i] = []int64{rand.Int63(), rand.Int63()}
		if rand.Intn(2) == 0 {
			res[i][0] *= -1
		}
		if rand.Intn(2) == 0 {
			res[i][1] *= -1
		}
	}
	return res
}

func TestMax(t *testing.T) {
	tests := gen_maxmin_tests()
	for _, v := range tests {
		mx := Max(v[0], v[1])
		if mx < v[0]+v[1]-mx {
			t.Error(
				"\nFor\n", v,
				"\nexpected\n", v[0]+v[1]-mx,
				"\ngot\n", v,
			)
			return
		}
	}
}

func TestMin(t *testing.T) {
	tests := gen_maxmin_tests()
	for _, v := range tests {
		mn := Min(v[0], v[1])
		if mn > v[0]+v[1]-mn {
			t.Error(
				"\nFor\n", v,
				"\nexpected\n", v[0]+v[1]-mn,
				"\ngot\n", v,
			)
			return
		}
	}
}

/*
type Instruction struct {
	Todo       byte   // 1 if save file
	Name       string // where to save file
	Data_size  int64  // -1 if don't need
	Data_ind   int64  // -1 if don't need
	Sign       []byte
	Req_amount int64
}
*/

func gen_instructions() []Instruction {
	res := make([]Instruction, 1e5) // e5
	for i := 0; i < len(res); i++ {
		res[i].Todo = byte(rand.Intn(256))
		len_name := rand.Intn(50)
		res[i].Name = ""
		switch res[i].Todo {
		case 1:
			for j := 0; j < len_name; j++ {
				res[i].Name += string(byte(20 + rand.Intn(100)))
			}
		}
		res[i].Data_size = rand.Int63()
		res[i].Data_ind = rand.Int63()
		len_sign := 1 + rand.Intn(20)
		res[i].Sign = make([]byte, len_sign)
		for j := 0; j < len_sign; j++ {
			res[i].Sign[j] = byte(rand.Intn(256))
		}
		res[i].Req_amount = rand.Int63()
	}
	return res
}

func Cmp_slices[T Number](a, b []T) bool { // true if equal, otherwise false
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

func cmp_instructions(i1, i2 *Instruction) bool { // true if equal, otherwise false
	if (*i1).Todo != (*i2).Todo {
		return false
	}
	if (*i1).Name != (*i2).Name {
		return false
	}
	if (*i1).Data_size != (*i2).Data_size {
		return false
	}
	if (*i1).Data_ind != (*i2).Data_ind {
		return false
	}
	if !Cmp_slices((*i1).Sign, (*i2).Sign) {
		return false
	}
	if (*i1).Req_amount != (*i2).Req_amount {
		return false
	}
	return true
}

func TestInstruction_code_decode(t *testing.T) {
	tests := gen_instructions()
	for _, v := range tests {
		inst_in_byte_arr, err := Parse_instruction(v)
		if err != nil {
			t.Error(
				"\nGot error\n", err,
				"\nfor\n", v.Todo, v.Name, v.Data_size, v.Data_ind, v.Sign, v.Req_amount,
			)
			return
		}
		parsed_inst := Parse_instruction_request(&inst_in_byte_arr, 0)
		if !cmp_instructions(&v, &parsed_inst) {
			t.Error(
				"\nParse instruction is not equal to base\n",
				"\nfor\n", v.Todo, v.Name, v.Data_size, v.Data_ind, v.Sign, v.Req_amount,
				"\ngot\n", inst_in_byte_arr,
				"\nand converted to\n",
				parsed_inst.Todo, parsed_inst.Name, parsed_inst.Data_size, parsed_inst.Data_ind, parsed_inst.Sign, parsed_inst.Req_amount,
			)
			return
		}
	}
}

func gen_int64_and_array_size() [][]int64 {
	res := make([][]int64, 1e5+1)
	for i := 0; i < len(res); i++ {
		res[i] = []int64{int64(rand.Intn(10)), rand.Int63()}
	}
	res[1e5] = []int64{1, 1e18}
	return res
}

func TestInt64_to_array_fixed_code_decode(t *testing.T) {
	tests := gen_int64_and_array_size()
	for _, v := range tests {
		int64_in_byte_arr, err := Int64_to_byte_arr_fixed_size(byte(v[0]), v[1])
		neg_number, max_arr_size_too_big, max_arr_size_too_small := false, false, false
		if v[1] < 0 {
			neg_number = true
		}
		if v[0] > Max_reqs_size {
			max_arr_size_too_big = true
		}
		power := int64(1)
		for i := 0; i < int(v[0])-1; i++ {
			power *= 256
		}
		if power < v[1] {
			max_arr_size_too_small = true
		}

		if neg_number && err == Negative_Number {
			continue
		}
		if max_arr_size_too_big && err == Max_arr_size_too_big {
			continue
		}
		if max_arr_size_too_small && err == Max_arr_size_too_small {
			continue
		}
		if err != nil {
			t.Error(
				"\nGot error\n", err,
				"\nfor", v[1], "\nwith size\n", v[0],
			)
			return
		}
		if len(int64_in_byte_arr) != int(v[0]) {
			t.Error(
				"\nArray should have size\n", v[0],
				"\nbut has\n", len(int64_in_byte_arr),
			)
			return
		}
		parsed_int64 := Byte_array_to_int64_fixed(int64_in_byte_arr)
		if v[1] != parsed_int64 {
			t.Error(
				"\nParsed int64 is not equal to base\n",
				"\nfor\n", v[1], "\nwith size\n", v[0],
				"\ngot\n", int64_in_byte_arr,
				"\nand converted to\n", parsed_int64,
			)
			return
		}
	}
}

func TestInt64_to_array_code_decode(t *testing.T) {
	tests := gen_int64_and_array_size()
	for _, v := range tests {
		int64_in_byte_arr, err := Int64_to_byte_arr(v[1])
		if err != nil {
			neg_number, too_big_number := false, false
			if v[1] < 0 {
				neg_number = true
			}

			power := int64(1)
			for i := 0; i < 7; i++ {
				power *= 256
			}
			if v[1]/power > 255 {
				too_big_number = true
			}

			if neg_number && err == Negative_Number {
				continue
			}
			if too_big_number && err == Input_number_too_big {
				continue
			}
			t.Error(
				"\nGot error\n", err,
				"\nfor\n", v[1], "\nwith size\n", v[0],
			)
			return
		}
		parsed_int64 := Byte_array_to_int64(int64_in_byte_arr)
		if v[1] != parsed_int64 {
			t.Error(
				"\nParsed int64 is not equal to base\n",
				"\nfor\n", v[1], "\nwith size\n", v[0],
				"\ngot\n", int64_in_byte_arr,
				"\nand converted to\n", parsed_int64,
			)
			return
		}
	}
}

type text_params struct {
	text      *[]byte
	num_parts int
	length    int
}

func gen_text_and_params() []text_params {
	result := make([]text_params, 1e5)
	for i := 0; i < len(result); i++ {
		var arr []byte
		if i < 100 {
			arr = gen_byte_arr(rand.Intn(10) + 1)
		} else if i < 110 {
			arr = gen_byte_arr(rand.Intn(1e6) + 1)
		} else {
			arr = gen_byte_arr(rand.Intn(1e3) + 1)
		}
		divs := make([]int, 0)
		for j := 1; j*j <= len(arr); j++ {
			if len(arr)%j == 0 {
				divs = append(divs, j)
			}
		}
		num_parts := divs[rand.Intn(len(divs))]
		result[i] = text_params{&arr, num_parts, len(arr) / num_parts}
	}
	return result
}

func TestDivide_text_into_parts(t *testing.T) {
	for _, v := range gen_text_and_params() {
		codec, result := Divide_text_into_parts(v.text, v.num_parts, v.num_parts*v.length)
		data_res, err := Get_byte_data(&result, &codec, v.num_parts*v.length)
		if err != nil {
			t.Error(err)
			return
		}
		if !Cmp_slices(data_res, *v.text) {
			addition := ""
			if len(data_res) <= 10 {
				addition = fmt.Sprintf("\nWith input \n%v, \noutput \n%v \ncodec \n%v",
					(*v.text), data_res, codec)
			}
			t.Error(
				"\nData after dividing and merging changed\n",
				"\ntext len:\n", len(*v.text),
				"\nnum_parts:\n", v.num_parts,
				"\npart_length:\n", v.length,
				addition,
			)
			return
		}
	}
}
