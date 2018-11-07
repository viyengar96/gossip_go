package main

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/sync/semaphore"
)

const numNodes = 8

var p_sem = semaphore.NewWeighted(1)

var hb_counter_t uint64
var hb_send_t uint64
var node_fail_t uint64
var shutdown_t uint64

type heartbeat struct {
	id   int
	hb   uint64
	time uint64
}

type hbTable struct {
	table []heartbeat
	id int
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: ./gossip_go <hb counter T> <hb send T> <node fail T>")
		os.Exit(1)
	}

	hb_counter_t_signed, cntr_err := strconv.ParseInt(os.Args[1], 10, 64)
	if cntr_err != nil {
		fmt.Println(cntr_err)
		os.Exit(1)
	}
	hb_send_t_signed, send_err := strconv.ParseInt(os.Args[2], 10, 64)
	if send_err != nil {
		fmt.Println(send_err)
		os.Exit(1)
	}
	node_fail_t_signed, nfail_err := strconv.ParseInt(os.Args[3], 10, 64)
	if nfail_err != nil {
		fmt.Println(nfail_err)
		os.Exit(1)
	}

	//Cast signed vars returned from strconv to unsigned
	hb_counter_t = uint64(hb_counter_t_signed)
	hb_send_t = uint64(hb_send_t_signed)
	node_fail_t = uint64(node_fail_t_signed)
	shutdown_t = 30

	fmt.Println(hb_counter_t, " ", hb_send_t, " ", node_fail_t)

	var gossip_channels [numNodes]chan heartbeat
	var print_channel = make(chan hbTable, numNodes+1)
	for i := range gossip_channels {
		gossip_channels[i] = make(chan heartbeat)
	}

	var connections = make(map[int][]int)

	for i := 0; i < numNodes; i++ {
		var v1, v2 int
		if i == 0 {
			v1 = numNodes - 1
		} else {
			v1 = i - 1
		}
		v2 = (i + 1) % numNodes
		connections[i] = []int{v1, v2}
	}

	for i := range connections {
		// fmt.Println("here");
		go spinUpNode(i, connections[i], gossip_channels, print_channel)
	}

	for {
		if len(print_channel) > 0 {
			p_val := <-print_channel
			fmt.Println("\n---------------------\n ");
			fmt.Println("Node: ", p_val.id, "\n ");
			fmt.Println("Table: \n", p_val.table);
		}
	}
}

func spinUpNode(id int, neighbors []int, channel [numNodes]chan heartbeat, p_channel chan hbTable) {
	//Vars to keep track of current node's heatbeat and local time
	table := []heartbeat{};
	
	var my_hb uint64;
	my_hb = 0;

	var local_fail_t = node_fail_t;
	var local_send_t = hb_send_t;
	var local_counter_t = hb_counter_t;

	//Vars for conditionals
	var local_t uint64
	local_t = 0 //Used to keep track of current node's local time

	for {
		//Simulate shutdown after x seconds
		if node_fail_t == local_t {
			local_fail_t += node_fail_t;
		}

		//Send current heartbeat to both neighbors
		if hb_send_t == local_t {
			local_send_t += hb_send_t;
		} else { //Recv current heartbeat from both neighbors

		}

		//Increment heartbeat periodically
		if local_counter_t == local_t {
			my_hb ++;
			local_counter_t += hb_counter_t;
		}

		local_t++ //Increment local time
	}
}

// func sendHB(id int)
