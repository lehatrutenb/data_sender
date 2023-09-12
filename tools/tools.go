package tools

import (
	"errors"
	"math/rand"
)

const Data_size_len = byte(7)
const Max_reqs_size = 8 // 255^7

type Number interface {
	~byte | ~int | ~int8 | ~int16 | ~int32 | ~int64
}

func Max[T Number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Min[T Number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

type Instruction struct {
	Todo       byte   // 1 if save file
	Name       string // where to save file
	Data_size  int64  // -1 if don't need
	Data_ind   int64  // -1 if don't need
	Sign       []byte
	Req_amount int64
}

func Parse_instruction_fixed_size(inst Instruction) ([]byte, error) {
	res := make([]byte, 0)
	res = append(res, inst.Todo)
	switch inst.Todo {
	case 1:
		res = append(append(res, byte(len(inst.Name))), []byte(inst.Name)...)
	}
	if inst.Data_size != -1 {
		data_size_byte_arr, err := Int64_to_byte_arr_fixed_size(Data_size_len, inst.Data_size)
		if err != nil {
			return []byte{}, err
		}
		res = append(res, append([]byte{Data_size_len}, data_size_byte_arr...)...)
	}
	if inst.Data_ind != -1 {
		data_ind_byte_arr, err := Int64_to_byte_arr_fixed_size(Data_size_len, inst.Data_ind)
		if err != nil {
			return []byte{}, err
		}
		res = append(res, append([]byte{Data_size_len}, data_ind_byte_arr...)...)
	}
	if len(inst.Sign) != 0 {
		res = append(res, append([]byte{byte(len(inst.Sign))}, inst.Sign...)...)
	}
	if inst.Req_amount != -1 {
		req_amount_byte_arr, err := Int64_to_byte_arr_fixed_size(Max_reqs_size, inst.Req_amount)
		if err != nil {
			return []byte{}, err
		}
		res = append(res, append([]byte{Max_reqs_size}, req_amount_byte_arr...)...)
	}
	res = append(res, 0)

	return res, nil
}

func Parse_instruction(inst Instruction) ([]byte, error) {
	res := make([]byte, 0)
	res = append(res, inst.Todo)
	switch inst.Todo {
	case 1:
		res = append(append(res, byte(len(inst.Name))), []byte(inst.Name)...)
	}
	if inst.Data_size != -1 {
		data_size_byte_arr, err := Int64_to_byte_arr(inst.Data_size)
		if err != nil {
			return []byte{}, err
		}
		res = append(res, data_size_byte_arr...)
	} else {
		res = append(res, 0)
	}

	if inst.Data_ind != -1 {
		data_ind_byte_arr, err := Int64_to_byte_arr(inst.Data_ind)
		if err != nil {
			return []byte{}, err
		}
		res = append(res, data_ind_byte_arr...)
	} else {
		res = append(res, 0)
	}

	if len(inst.Sign) != 0 {
		res = append(res, append([]byte{byte(len(inst.Sign))}, inst.Sign...)...)
	} else {
		res = append(res, 0)
	}

	if inst.Req_amount != -1 {
		req_amount_byte_arr, err := Int64_to_byte_arr(inst.Req_amount)
		if err != nil {
			return []byte{}, err
		}
		res = append(res, req_amount_byte_arr...)
	} else {
		res = append(res, 0)
	}

	res = append(res, 0)

	return res, nil
}

func Parse_instruction_request(buffer *[]byte, left int) (inst Instruction, left_ int) {
	inst.Todo = (*buffer)[left]
	name_size := 0
	left++
	switch inst.Todo {
	case 1:
		name_size = int((*buffer)[left])
		inst.Name = string((*buffer)[left+1 : left+1+name_size])
		left = left + 1 + name_size
	}

	var data_size int
	if (*buffer)[left] != 0 {
		data_size = int((*buffer)[left])
		inst.Data_size = Byte_array_to_int64((*buffer)[left : left+1+data_size])
		left = left + 1 + data_size
	} else {
		left++
	}

	if (*buffer)[left] != 0 {
		index_size := int((*buffer)[left])
		inst.Data_ind = Byte_array_to_int64((*buffer)[left : left+1+index_size])
		left = left + 1 + index_size
	} else {
		left++
	}

	if (*buffer)[left] != 0 {
		sign_size := int((*buffer)[left])
		inst.Sign = (*buffer)[left+1 : left+1+sign_size]
		left = left + 1 + sign_size
	} else {
		left++
	}
	if (*buffer)[left] != 0 {
		req_amount_size := int((*buffer)[left])
		inst.Req_amount = Byte_array_to_int64((*buffer)[left : left+1+req_amount_size])
		left = left + 1 + req_amount_size
	} else {
		left++
	}
	left++

	return inst, left
}

type Skin_cl struct {
	Sign       []byte
	Req_amount []byte
	Index      []byte
}

type Skin_ser struct {
	Sign       []byte
	Req_amount int64
	Index      int64
}

func Lenb(p []byte) byte {
	if p != nil {
		return byte(len(p))
	}
	return 0
}

func (sk *Skin_cl) Full_len() byte {
	return Lenb(sk.Sign) + Lenb(sk.Index) + Lenb(sk.Req_amount)
}

func Add_skin(buffer *[]byte, sk *Skin_cl) {
	byte_skin := make([]byte, 0, sk.Full_len())

	byte_skin = append(append(byte_skin, Lenb(sk.Sign)), sk.Sign...)
	byte_skin = append(append(byte_skin, Lenb(sk.Req_amount)), sk.Req_amount...)
	byte_skin = append(append(byte_skin, Lenb(sk.Index)), sk.Index...)
	*buffer = append(byte_skin, *buffer...)
}

var Negative_Number = errors.New("Cant convert negative number to array")
var Max_arr_size_too_big = errors.New("max_arr_size is too big - max power is bigger then int64")
var Max_arr_size_too_small = errors.New("max_arr_size is too small - power * 255 is less then number")
var Input_number_too_big = errors.New("can't convert number to array - log256(x) should be less then 8")

func Int64_to_byte_arr_fixed_size(max_arr_size byte, x int64) ([]byte, error) {
	if x < 0 {
		return []byte{}, Negative_Number
	}
	if max_arr_size > Max_reqs_size {
		return []byte{}, Max_arr_size_too_big
	}
	if max_arr_size == 0 {
		return []byte{}, Max_arr_size_too_small
	}
	num_reqs := make([]byte, max_arr_size)
	power := int64(1)
	for i := byte(0); i < max_arr_size-1; i++ {
		power *= 256
	}

	if x/power > 255 {
		return []byte{}, Max_arr_size_too_small
	}

	for i := int(max_arr_size) - 1; i >= 0; i-- {
		num_reqs[i] = byte(x / power)
		x -= (x / power) * power
		power /= 256
	}
	return num_reqs, nil
}

func Int64_to_byte_arr(x int64) ([]byte, error) {
	if x < 0 {
		return []byte{}, Negative_Number
	}
	num_reqs := make([]byte, 0)
	power := int64(1)
	kol := byte(0)
	for x/power > 255 {
		kol++
		power *= 256
		if kol == 7 && x/power > 255 {
			return []byte{}, Input_number_too_big
		}
	}

	for power >= 1 {
		num_reqs = append([]byte{byte(x / power)}, num_reqs...)
		x -= (x / power) * power
		power /= 256
	}

	num_reqs = append([]byte{byte(len(num_reqs))}, num_reqs...)

	return num_reqs, nil
}

func Byte_array_to_int64_fixed(arr []byte) int64 {
	res := int64(0)
	power := int64(1)

	for i := 0; i < len(arr); i++ {
		res += power * int64(arr[i])
		power *= 256
	}

	return res
}

func Byte_array_to_int64(arr []byte) int64 {
	res := int64(0)
	power := int64(1)

	for i := 1; i < len(arr); i++ {
		res += power * int64(arr[i])
		power *= 256
	}

	return res
}

func Write_chan(ch chan byte) {
	if ch != nil {
		ch <- 1
	}
}

func Divide_arr_into_parts(num_parts, length int) [][]int {
	var arr = make([]int, length)

	for i := 0; i < length; i++ {
		arr[i] = i
	}

	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})

	result := make([][]int, num_parts)

	pixels_per_part := Max(1, (length+num_parts-1)/num_parts)

	ind := 0
	now := 0
	for _, ch := range arr {
		if now == pixels_per_part {
			now = 0
			ind++
			ind %= num_parts
		}
		result[ind] = append(result[ind], ch)
		now++
	}

	return result
}

func Divide_text_into_parts(text *[]byte, num_parts int, length int) ([][]int, [][]byte) {
	codec := Divide_arr_into_parts(num_parts, length)

	result := make([][]byte, num_parts)
	for i := 0; i < num_parts; i++ {
		for _, ind := range codec[i] {
			result[i] = append(result[i], (*text)[ind])
		}
	}

	return codec, result
}

func Send_byte_data(data []byte, num_parts int, to string) [][]int {
	codec, result := Divide_text_into_parts(&data, num_parts, len(data))
	if len(result) != 0 {
	}
	return codec
}

func Get_byte_data(result *[][]byte, codec *[][]int, length int) ([]byte, error) {
	if len(*result) != len(*codec) {
		return nil, errors.New("Result and codec should have same length")
	}

	data := make([]byte, length)
	for i := 0; i < len(*codec); i++ {
		for j, ind := range (*codec)[i] {
			data[ind] = (*result)[i][j]
		}
	}

	return data, nil
}
