package main;

import (
	"fmt"
	"golang.org/x/sync/semaphore"
)

const numNodes = 8;

var p_sem = semaphore.NewWeighted(1);

type heartbit struct {
	id int
	hb uint64
	time uint64
}


func main() {
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
	
	local_t := 0;
	for {
		if local_t == 10  {
			p_channel <- my_hb
		}
		local_t ++;
	}
}