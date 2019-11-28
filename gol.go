package main

import (
	"fmt"
	"strconv"
	"strings"
)

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p golParams, d distributorChans, alive chan []cell) {

	// Create the 2D slice to store the world.
	world := make([][]byte, p.imageHeight)
	for i := range world {
		world[i] = make([]byte, p.imageWidth)
	}

	// Create the 2D slice to store the new world.
	new_world := make([][]byte, p.imageHeight)
	for i := range world {
		new_world[i] = make([]byte, p.imageWidth)
	}

	// Create the 2D slice to store the world.
	source_slice_world := make([][]byte, p.imageHeight)
	for i := range source_slice_world {
		source_slice_world[i] = make([]byte, p.imageWidth)
	}

	// Create the 2D slice to store the new world.
	target_slice_world := make([][]byte, p.imageHeight)
	for i := range target_slice_world {
		target_slice_world[i] = make([]byte, p.imageWidth)
	}

	// Request the io goroutine to read in the image with the given filename.
	d.io.command <- ioInput
	d.io.filename <- strings.Join([]string{strconv.Itoa(p.imageWidth), strconv.Itoa(p.imageHeight)}, "x")

	// The io goroutine sends the requested image byte by byte, in rows.
	for y := 0; y < p.imageHeight; y++ {
		for x := 0; x < p.imageWidth; x++ {
			val := <-d.io.inputVal
			if val != 0 {
				fmt.Println("Alive cell at", x, y)
				world[y][x] = val
			}
		}
	}

	// Calculate the new state of Game of Life after the given number of turns.
	for turns := 0; turns < p.turns; turns++ {
		splitUpBoard(world, p)
		golLogic(world, new_world, p, d)
		//putboardbacktogether
		d.io.command <- ioOutput
		d.io.filename<-strings.Join([]string{strconv.Itoa(p.imageWidth), strconv.Itoa(p.imageHeight),strconv.Itoa(p.threads),strconv.Itoa(p.turns)} , "x")
		swap(world, new_world, p, d)
	}


	// Create an empty slice to store coordinates of cells that are still alive after p.turns are done.
	var finalAlive []cell
	// Go through the world and append the cells that are still alive.
	for y := 0; y < p.imageHeight; y++ {
		for x := 0; x < p.imageWidth; x++ {
			if world[y][x] != 0 {
				finalAlive = append(finalAlive, cell{x: x, y: y})
			}
		}
	}

	// Make sure that the Io has finished any output before exiting.
	d.io.command <- ioCheckIdle
	<-d.io.idle

	// Return the coordinates of cells that are still alive.
	alive <- finalAlive
}

func swap(world [][]byte, new_world [][]byte, p golParams, d distributorChans){
	for y := 0; y < p.imageHeight; y++ {
		for x := 0; x < p.imageWidth; x++ {
			world[y][x] = new_world[y][x]
			d.io.outputVal <- world[y][x]
		}
	}
}

func golLogic(world [][]byte, new_world [][]byte, p golParams, d distributorChans){
	for y := 0; y < p.imageHeight; y++ {
		for x := 0; x < p.imageWidth; x++ {

			sum := 0
			maxHeight := p.imageHeight - 1
			maxWidth := p.imageWidth - 1

			if (y == 0) || (y == maxHeight) || (x == 0) || (x == maxWidth){
				yplus  := y + 1
				yminus := y - 1
				xplus  := x + 1
				xminus := x - 1

				if (y == 0){
					yminus = maxHeight
				}
				if (y == maxHeight){
					yplus = 0
				}
				if (x == 0){
					xminus = maxWidth
				}
				if (x == maxWidth){
					xplus = 0
				}


				if world[yminus][xminus] == 0xFF	{sum++}
				if world[yminus][x] == 0xFF			{sum++}
				if world[yminus][xplus] == 0xFF		{sum++}

				if world[y][xminus] == 0xFF			{sum++}
				if world[y][xplus] == 0xFF			{sum++}

				if world[yplus][xminus] == 0xFF		{sum++}
				if world[yplus][x] == 0xFF			{sum++}
				if world[yplus][xplus] == 0xFF		{sum++}
			}else{

				for horizontal := -1; horizontal < 2; horizontal++ {
					if world[y-1][x+horizontal] == 0xFF	{sum++}
					if world[y+1][x+horizontal] == 0xFF	{sum++}
				}
				if world[y][x-1] == 0xFF		{sum++}
				if world[y][x+1] == 0xFF		{sum++}
			}

			//GOL rules
			if (sum < 2) && (world[y][x] == 0xFF){
				new_world[y][x] = 0x00
			}

			if ((sum == 2) || sum == 3 ) && (world[y][x] == 0xFF){
				new_world[y][x] = 0xFF
			}

			if (sum > 3) && (world[y][x] == 0xFF) {
				new_world[y][x] = 0x00
			}

			if (sum == 3) && (world[y][x] == 0x00) {
				new_world[y][x] = 0xFF
			}
		}
	}
}

func print_out_board(current_world [][]byte, p golParams, turns int){
	println(turns)
	for y := 0; y < p.imageHeight; y++ {
		for x := 0; x < p.imageWidth; x++ {
			if current_world[y][x] == 0xFF {
				print("0 ")
			}else{
				print("- ")
			}
		}
		print("\n")
	}
	print("\n \n")
}

func worker(p golParams, d distributorChans){





}

func splitUpBoard(world [][]byte, p golParams){
	slice_height := (p.imageHeight/ p.threads) + 2 //height of slices we are sending

	for i := 0; i < p.threads; i++ { //for each thread to send
		start := i*slice_height //start of the thread
		top := start - 1 //edge cases
		if start == 0 {
			top = maxHeight
		}
		bottom := start + slice_height
		if start == p.imageHeight - workerHeight {
			bottom = 0
		}
	}
	sendRows(world, p, top, bottom)
}

func sendRows(world [][] byte, p golParams, startRow int, endRow int, d distributorChans){
	d.io.channel_thread <- world[startRow][x] //send top halo

	middleStart := startRow + 1
	middleEnd := endRow - 1

	for s := middleStart; s < middleEnd; s++{ //send middle halo
		d.io.channel_thread <- world[s][x]
	}

	d.io.channel_thread <- world[endRow][x] //send bottom halo
}

func recieveRows(target_slice_world [][] byte, world [][] byte, p golParams, startRow int, endRow Int){
	middleStart := startRow + 1
	middleEnd := endRow - 1
	for s := middleStart; s < middleEnd; s++{
		world[y][x] = target_slice_world[s][x]
	}
}