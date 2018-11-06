package main;

import (
	"fmt"
	"os"
	"golang.org/x/sync/semaphore"
	"strconv"
)

const numNodes = 8;

var p_sem = semaphore.NewWeighted(1);

var hb_counter_T int64;
var hb_send_T int64;
var node_fail_T int64;
var err error;
var shutdown_T int64;

type heartbit struct {
	id int
	hb uint64
	time uint64
}

func main() {
	if(len(os.Args) != 4){
		fmt.Println("Usage: ./gossip_go <hb counter T> <hb send T> <node fail T>");
		os.Exit(1);
	}

	hb_counter_T, err = strconv.ParseInt(os.Args[1], 10, 64);
	if err != nil {
		fmt.Println(err);
		os.Exit(1);
	}
	hb_send_T, err = strconv.ParseInt(os.Args[2], 10, 64);
	if err != nil {
		fmt.Println(err);
		os.Exit(1);
	}
	node_fail_T, err = strconv.ParseInt(os.Args[3], 10, 64);
	if err != nil {
		fmt.Println(err);
		os.Exit(1);
	}

	shutdown_T = 30;

	fmt.Println(hb_counter_T, " ", hb_send_T, " ", node_fail_T);

	var gossip_channels [numNodes]chan heartbit;
	var print_channel = make(chan heartbit, numNodes+1);
	for i := range gossip_channels { 
		gossip_channels[i] = make(chan heartbit);
	}

	var connections = make(map[int][]int);

	for i := 0; i < numNodes; i++ {
		var v1, v2 int;
		if(i == 0){
			v1 = numNodes - 1;
		} else {
			v1 = i - 1;
		}
		v2 = (i+1)%numNodes
		connections[i] = []int{v1, v2};
	}

	for i := range connections {
		// fmt.Println("here");
		go spinUpNode(i, connections[i], gossip_channels, print_channel);
	}

	for {
		if len(print_channel) > 0 {
			p_val := <- print_channel;
			fmt.Println(p_val);
		}
	}
}

func spinUpNode(id int, neighbors []int, channel [numNodes]chan heartbit, p_channel chan heartbit) {
	// for !p_sem.TryAcquire(1){}
	// fmt.Println(id);
	// p_sem.Release(1);
	var my_hb heartbit;
	my_hb.id = id;
	my_hb.hb = 0;
	my_hb.time = 0;
	
	var local_t uint64;
	local_t = 0;

	for {
		if local_t == 10  {
			p_channel <- my_hb
		}
		local_t ++;
	}
}