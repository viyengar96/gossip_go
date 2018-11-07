package main

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/sync/semaphore"
)

const shutdown_duration uint64 = 4000000000 //duration that node remains shutdown
const numNodes = 8
const numChannels = numNodes *2;

var p_sem = semaphore.NewWeighted(1)

var hb_counter_t uint64
var hb_send_t uint64
var shutdown_t uint64

var fail_t uint64 = 3000;		//Mark node for removal from hb table if local_t - table[i].time >= fail_t


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

	for {
		if len(print_channel) > 0 {
			p_val := <-print_channel
			fmt.Println("\n---------------------\n ");
			fmt.Println("Node: ", p_val.id, "\n ");
			fmt.Println("Table: \n", p_val.table);
		}
	}
}

func spinUpNode(id int, neighbors []int, channels [numChannels]chan []heartbeat, p_channel chan hbTable, master_table []heartbeat) {
	//Vars to keep track of current node's heatbeat and local time
	table := master_table;
	fail_array := []bool{false, false, false, false, false, false, false, false};
	
	var my_hb uint64;
	my_hb = 0;

	var local_shutdown_t = shutdown_t;
	var local_send_t = hb_send_t;
	var local_counter_t = hb_counter_t;

	//Vars for conditionals
	var local_t uint64
	local_t = 0 //Used to keep track of current node's local time

	for {
		//Simulate shutdown after x seconds
		if local_shutdown_t == local_t {
			var pTable hbTable;
			pTable.table = table;
			pTable.id = id;
			// fmt.Printf("SHUTDOWN %d\n", id);
			p_channel <- pTable;
			return;
		}

		//Send current heartbeat to both neighbors
		if hb_send_t == local_t {
			// fmt.Printf("SEND 1: %d\n", id);
			channels[neighbors[0]] <- table;
			// sendHB(table, channels[2]);
			// fmt.Printf("SEND 2: %d \n", id);
			channels[neighbors[2]] <- table;
			local_send_t += hb_send_t;
		}
		//Recv current heartbeat from both neighbors
		if len(channels[neighbors[1]]) > 0 {
			// fmt.Printf("RECV 1 %d \n", id);
			newTable := <- channels[neighbors[1]];
			updateTable(newTable, table, local_t, fail_array);
		}
			
		if len(channels[neighbors[3]]) > 0 {
			// fmt.Printf("RECV 2 %d \n", id);
			newTable := <- channels[neighbors[3]];
			updateTable(newTable, table, local_t, fail_array);
		}

		table = checkTable(table, fail_array, local_t);

		//Increment heartbeat periodically
		if local_counter_t == local_t {
			// fmt.Printf("HB INC %d \n", id);
			my_hb ++;
			for i := range table {
				if table[i].id == id {
					table[i].hb = my_hb;
				}
			}
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
				} else if time > oldTable[j].time {
					oldTable[i].hb = newTable[j].hb;
					oldTable[i].time = time;
				}
			}
		}
	}
}

func checkTable(table []heartbeat, fail_array []bool, time uint64) []heartbeat{
	newTable := []heartbeat{}
	for i := range table {
		// fmt.Println(i);
		susNode := table[i].id;
		//If array has been marked and time cleanup has been reached
		if fail_array[susNode] && (time - table[i].time) >= fail_t*2{ 
			//Remove node from table
			// newTable = removeTableEntry(table, i);
			// pendingRemoval = append(pendingRemoval, i);
			fmt.Printf("NODE %d IS DEAD", susNode);
			continue;
		}
		if (time - table[i].time) >= fail_t {
			//mark for failure
			fail_array[susNode] = true;
			fmt.Printf("NODE %d IS MARKED", susNode);
		}
		newTable = append(newTable, table[i]);
		// fmt.Println(susNode);
	}
	
	
	return newTable;
}

func removeTableEntry(oldTable []heartbeat, i int) []heartbeat {
	return append(oldTable[:i], oldTable[i+1:]...)
}

// func shutDownNode() {
// 	var local_sd_duration uint64
// 	local_sd_duration = shutdown_duration

// 	//Shutdown for as long as the shutdown duraiton time
// 	for local_sd_duration > 0 {
// 		local_sd_duration--
// 	}
// }
