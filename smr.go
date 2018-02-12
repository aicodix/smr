/*
SMR technology drive management torture test

Copyright 2018 Ahmet Inan <inan@aicodix.de>
*/

package main

import (
	"os"
	"fmt"
	"flag"
	"math"
	"time"
	"unsafe"
	"strconv"
	"syscall"
	"io/ioutil"
	"math/rand"
)

func AlignedBytes(size, align uint) []byte {
	buf := make([]byte, size + align, size + align)
	ptr := unsafe.Pointer(&buf[0])
	off := align - uint(uintptr(ptr) % uintptr(align))
	return buf[off : off + size]
}

func BlockSize(file *os.File) (int, error) {
	const BLKSSZGET = 0x1268
	var ssz int
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), BLKSSZGET, uintptr(unsafe.Pointer(&ssz)))
	if errno != 0 {
		return 0, &os.SyscallError{"BLKSSZGET", errno}
	}
	return ssz, nil
}

func ChooseScheduler(name, sched string) error {
	return ioutil.WriteFile("/sys/block/" + name + "/queue/scheduler", []byte(sched), 0)
}

func HumanReadable(val float64) string {
	prefix := ""
	for _, r := range "KMGT" {
		if math.Abs(val) < 1024.0 {
			break
		}
		val /= 1024.0
		prefix = string(r) + "i"
	}
	prec := 3
	for t := math.Abs(val); t >= 1.0 && prec > 0; t /= 10.0 {
		prec--
	}
	return strconv.FormatFloat(val, 'f', prec, 64) + " " + prefix
}

func main() {
	var human bool
	flag.BoolVar(&human, "human", false, "make output human readable")
	var random bool
	flag.BoolVar(&random, "random", false, "write to random positions")
	var garbage bool
	flag.BoolVar(&garbage, "garbage", false, "write random data")
	var direct bool
	flag.BoolVar(&direct, "direct", false, "use direct io")
	var sync bool
	flag.BoolVar(&sync, "sync", false, "sync each write")
	var chunk uint
	flag.UintVar(&chunk, "chunk", 512, "chunk size")
	var total uint64
	flag.Uint64Var(&total, "total", 1048576, "total bytes to write")
	var device string
	flag.StringVar(&device, "device", "sdb", "block device name")
	var scheduler string
	flag.StringVar(&scheduler, "scheduler", "noop", "choose queue scheduler")
	flag.Parse()

	err := ChooseScheduler(device, scheduler)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error while setting scheduler:", err)
		return
	}
	flags := syscall.O_RDWR
	if direct {
		flags |= syscall.O_DIRECT
	}
	if sync {
		flags |= syscall.O_SYNC
	}
	file, err := os.OpenFile("/dev/" + device, flags, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error while opening device:", err)
		return
	}

	last, err := file.Seek(0, 2)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error while seeking to the end:", err)
		return
	}
	first, err := file.Seek(0, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error while seeking to begining:", err)
		return
	}
	size := last - first
	fmt.Fprintln(os.Stderr, "device size in bytes:", size)
	chunks := size / int64(chunk)
	fmt.Fprintln(os.Stderr, "number of", chunk, "byte sized chunks:", chunks)

	var buf []byte
	if direct {
		ssz, err := BlockSize(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error while getting sector size:", err)
			return
		}
		if ssz < 0 {
			fmt.Fprintln(os.Stderr, "BLKSSZGET returned negative number")
			return
		}
		block := uint(ssz)
		fmt.Fprintln(os.Stderr, "logical block size:", block)
		buf = AlignedBytes(chunk, block)
	} else {
		buf = make([]byte, chunk, chunk)
	}

	for sum, pos, count := uint64(chunk), int64(0), 0; sum <= total; sum, pos, count = sum + uint64(chunk), pos + int64(chunk), count + 1 {
		if random {
			pos = int64(chunk) * rand.Int63n(chunks)
			r, err := file.Seek(pos, 0)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error while seeking:", err)
				return
			}
			if r != pos {
				fmt.Fprintln(os.Stderr, "wrong seek result:", r)
				return
			}
		}
		if garbage {
			rand.Read(buf)
		}
		start := time.Now()
		n, err := file.Write(buf)
		finish := time.Now()
		elapsed := finish.Sub(start)
		bps := float64(n) / elapsed.Seconds()
		if human {
			fmt.Printf("%6d @ %011x %sB written %sB/s\n", count, pos, HumanReadable(float64(sum)), HumanReadable(bps))
		} else {
			fmt.Println(count, pos, sum, bps)
		}
		if err == nil && len(buf) != n {
			fmt.Fprintln(os.Stderr, "partial write")
			return
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "error while writing:", err)
			return
		}
	}
}

