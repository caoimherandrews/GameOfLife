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
			golLogic(world, new_world, p, d)
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
				sum := 0											//counts the number of living neighbours a cell has
				if (x>0) && (x<p.imageWidth-1) && (y>0) &&(y < p.imageWidth - 1) {

					//if we're not dealing with any edge cases...
					for i := -1; i < 2; i++ {
						if world[y-1][x+i] == 0xFF	{sum++}			//the first row
						if world[y+1][x+i] == 0xFF	{sum++}			//the last row
					}
					if world[y][x-1] == 0xFF		{sum++}			//the left middle pixel
					if world[y][x+1] == 0xFF		{sum++}			//the right middle pixel

				} else{								//if there's edge cases, we need to look at checking each neighbour's validity individually
					yplus  := y+1					//variables to store new wraparound values
					yminus := y-1
					xplus  := x+1
					xminus := x-1

					if (y == 0){ 					//if at top of image
						yminus = p.imageHeight -1 	//to account for indexing from zero
					}
					if (y==p.imageHeight-1){ 		//if at bottom of image
						yplus = 0
					}
					if (x==0){						//if at left of image
						xminus = p.imageWidth -1
					}
					if (x == p.imageWidth-1){		//if at right of image
						xplus = 0
					}

					//find alive_count here for wrap around cases
					//top row
					if world[yminus][xminus] == 0xFF	{sum++}		//top left
					if world[yminus][x] == 0xFF			{sum++}		//top middle
					if world[yminus][xplus] == 0xFF		{sum++}		//top right

					//middle pixels
					if world[y][xminus] == 0xFF			{sum++}		//middle left
					if world[y][xplus] == 0xFF			{sum++}		//middle right

					//bottom row
					if world[yplus][xminus] == 0xFF		{sum++}		//bottom left
					if world[yplus][x] == 0xFF			{sum++}		//bottom middle
					if world[yplus][xplus] == 0xFF		{sum++}		//bottom right

				}

				//add to new image representation
				//any live cell with fewer than 2 neighbours dies
				if (world[y][x] == 0xFF) && (sum < 2) {
					new_world[y][x] = 0X00
				}

				//any live cell with 2 or 3 live neighbours is unaffected
				if (world[y][x] == 0xFF) && ((sum == 2) || (sum == 3)) {
					new_world[y][x] = 0xFF
				}

				//any live cell with more than 3 live neighbours dies
				if (world[y][x] == 0xFF) && (sum > 3) {
					new_world[y][x] = 0X00
				}

				//any dead cell with exactly 3 neighbours comes alive
				if (world[y][x] == 0x00) && (sum == 3) { 			//if it's dead, and has 3 neighbours
					new_world[y][x] = 0xFF 							//colour changed to black
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