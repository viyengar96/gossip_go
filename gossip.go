package main

import (
	"fmt"
	"os"
	"strconv"
	_ "time"

	"golang.org/x/sync/semaphore"
)

const numNodes = 8
const numChannels = numNodes * 2;

var downCounter = numNodes;

var p_sem = semaphore.NewWeighted(1)

var hb_counter_t uint64
var hb_send_t uint64
var shutdown_t uint64

var fail_t uint64;		//Mark node for removal from hb table if local_t - table[i].time >= fail_t

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
	shutdown_t = uint64(node_fail_t_signed)
	fail_t = shutdown_t/3;

	fmt.Println(hb_counter_t, " ", hb_send_t, " ", shutdown_t)

	var gossip_channels [numChannels]chan []heartbeat
	for i := range gossip_channels {
		gossip_channels[i] = make(chan []heartbeat, 10000000)
	}

	var print_channel = make(chan hbTable, numNodes+1)

	var connections = make(map[int][]int)
	//Each node has 4 shared channels: {w1, r1, w2, r2}
	connections[0] = []int {0, 1, 15, 14};
	connections[1] = []int {2, 3, 1, 0};
	connections[2] = []int {4, 5, 3, 2};
	connections[3] = []int {6, 7, 5, 4};
	connections[4] = []int {8, 9, 7, 6};
	connections[5] = []int {10, 11, 9, 8};
	connections[6] = []int {12, 13, 11, 10};
	connections[7] = []int {14, 15, 13, 12};

	var master_table = []heartbeat{};
	for i := range connections {
		var temp_hb heartbeat;
		temp_hb.id = i;
		temp_hb.hb = 0;
		temp_hb.time = 0;
		master_table = append(master_table, temp_hb)
	}

	for i := range connections {
		go spinUpNode(i, connections[i], gossip_channels, print_channel, master_table)
	}

	for downCounter > 0{
		if len(print_channel) > 0 {
			p_val := <-print_channel
			fmt.Println("\n---------------------\n ");
			fmt.Println("Node: ", p_val.id, "\n ");
			fmt.Println("Table: ");
			for i := range p_val.table {
				fmt.Println(p_val.table[i])
			}

		}
	}
}

func spinUpNode(id int, neighbors []int, channels [numChannels]chan []heartbeat, p_channel chan hbTable, master_table []heartbeat) {
	//Vars to keep track of current node's heatbeat and local time
	table := master_table;
	fail_array := []bool{false, false, false, false, false, false, false, false};
	
	var my_hb uint64;
	my_hb = 0;

	var local_shutdown_t = 5*shutdown_t*uint64(id+1)*100;
	var local_send_t = hb_send_t;
	var local_counter_t = hb_counter_t;

	//Vars for conditionals
	var local_t uint64
	local_t = 0 //Used to keep track of current node's local time

	for {
		//Simulate shutdown after x seconds
		if local_shutdown_t == local_t {
			downCounter -= 1;
			return;
		}

		// if (local_send_t*5) == local_t {
		// }

		//Send current heartbeat to both neighbors
		if local_send_t == local_t {
			channels[neighbors[0]] <- table;
			channels[neighbors[2]] <- table;
			local_send_t += hb_send_t;
		}

		// time.Sleep(5 * time.Millisecond);
		//Recv current heartbeat from both neighbors
		if len(channels[neighbors[1]]) > 0 {
			newTable := <- channels[neighbors[1]];
			updateTable(newTable, table, local_t, fail_array);
		}
			
		if len(channels[neighbors[3]]) > 0 {
			newTable := <- channels[neighbors[3]];
			updateTable(newTable, table, local_t, fail_array);
		}

		table = checkTable(id, table, fail_array, local_t, p_channel);

		//Increment heartbeat periodically
		if local_counter_t == local_t {
			my_hb ++;
			for i := range table {
				if table[i].id == id {
					table[i].hb = my_hb;
					table[i].time = local_t;
				}
			}
			var pTable hbTable;
			pTable.table = table;
			pTable.id = id;
			p_channel <- pTable;
			local_counter_t += hb_counter_t;
		}

		local_t++ //Increment local time
	}
}

func updateTable(newTable []heartbeat, oldTable []heartbeat, time uint64, fail_array []bool) {
	for i := range oldTable {
		for j := range newTable {
			if oldTable[i].id == newTable[j].id {
				if newTable[j].hb > oldTable[i].hb {
					oldTable[i].hb = newTable[j].hb;
					oldTable[i].time = time;
				}
			}
		}
	}
}

func checkTable(id int, table []heartbeat, fail_array []bool, time uint64, p_channel chan hbTable) []heartbeat{
	newTable := []heartbeat{};
	validIndices := []int{};
	for i := range table {
		susNode := table[i].id;
		//If array has been marked and time cleanup has been reached
		if fail_array[susNode] && (time - table[i].time) >= fail_t*uint64(id+1)*200{ 
			//Remove node from table
			continue;
		}
		if (time - table[i].time) >= fail_t*uint64(id) {
			//mark for failure
			fail_array[susNode] = true;
		}
		validIndices = append(validIndices, i);
	}	

	for i := range validIndices {
		newTable = append(newTable, table[validIndices[i]])
	}
	
	return newTable;
}

func removeTableEntry(oldTable []heartbeat, i int) []heartbeat {
	return append(oldTable[:i], oldTable[i+1:]...)
}
