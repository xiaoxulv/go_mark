// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Generating random text: a Markov chain algorithm

Based on the program presented in the "Design and Implementation" chapter
of The Practice of Programming (Kernighan and Pike, Addison-Wesley 1999).
See also Computer Recreations, Scientific American 260, 122 - 125 (1989).

A Markov chain algorithm generates text by creating a statistical model of
potential textual suffixes for a given prefix. Consider this text:

	I am not a number! I am a free man!

Our Markov chain algorithm would arrange this text into this set of prefixes
and suffixes, or "chain": (This table assumes a prefix length of two words.)

	Prefix       Suffix

	"" ""        I
	"" I         am
	I am         a
	I am         not
	a free       man!
	am a         free
	am not       a
	a number!    I
	number! I    am
	not a        number!

To generate text using this table we select an initial prefix ("I am", for
example), choose one of the suffixes associated with that prefix at random
with probability determined by the input statistics ("a"),
and then create a new prefix by removing the first word from the prefix
and appending the suffix (making the new prefix is "am a"). Repeat this process
until we can't find any suffixes for the current prefix or we exceed the word
limit. (The word limit is necessary as the chain table may contain cycles.)

Our version of this program reads text from standard input, parsing it into a
Markov chain, and writes generated text to standard output.
The prefix and output lengths can be specified using the -prefix and -words
flags on the command-line.
*/
package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"strconv"
)

// Prefix is a Markov chain prefix of one or more words.
type Prefix []string

/*
 * Suffix is a struct that maintains every prefix's suffix word and its frequency
 */
type Suffix struct{
	word string
	frequency int
}

// String returns the Prefix as a string (for use as a map key).
func (p Prefix) String() string {
	return strings.Join(p, " ")
}

// Shift removes the first word from the Prefix and appends the given word.
func (p Prefix) Shift(word string) {
	copy(p, p[1:])
	p[len(p)-1] = word
}

/* Chain contains a map ("chain") of prefixes to a list of suffixes.
 * A prefix is a string of prefixLen words joined with spaces.
 * A suffix is a slice of struct Suffix. A prefix can have multiple suffixes.
 */
type Chain struct {
	chain map[string][]Suffix
	prefixLen int
}

// NewChain returns a new Chain with prefixes of prefixLen words.
func NewChain(prefixLen int) *Chain {
	return &Chain{make(map[string][]Suffix), prefixLen}
}
/*
 * Build reads text from the provided slice of inputfile
 * parses it into prefixes and suffixes that are stored in Chain.
 */
func (c *Chain) Build(inputFile []string) {
	n := len(inputFile)//number of input files
	var s [][]string = make([][]string, n)//nest slices to store content of input
	for i := range s{
		s[i] = make([]string, 0)
	}

	//for each input file
	for i := 0; i < n; i++{
		in, err := os.Open(inputFile[i])
		if err != nil {
			fmt.Println("Error: couldn’t open the file")
			os.Exit(3) 
		}

		scanner := bufio.NewScanner(in)
		scanner.Split(bufio.ScanWords)//split by white space get words 

		for scanner.Scan(){
			s[i] = append(s[i], scanner.Text())//each file gets a slice of words
		}
	}
	for i, _ := range s{
		p := make(Prefix, c.prefixLen)
		for j, get := range s[i]{//get word from slice

			key := p.String()
			/*
			* maps of structs: can’t change the value of a field in a 
		 	* struct that is in a map. solution: use a copy!!
			* be careful when it comes to slices of struct as value field in map 
			*/
			suf := c.chain[key]//a slice of suffix of key's
			var find bool = false
			for i, value := range suf{
				if value.word == get{//suffix exists in table, frequency++
					value.frequency++
					suf[i] = value
					find = true
				}
			}
			if (find != true){//suffix not exists in table, frequency = 1
				var newSuf Suffix
				newSuf.word = get
				newSuf.frequency = 1
				c.chain[key] = append(c.chain[key], newSuf)
			}
			p.Shift(s[i][j])
		}
	}
}
/*
 * WirteFreTable writes chain in to output file.
 * The format should be prefix Suffix{word frequency}.
 * First line inpliews the prefixLen.
 */
func (c *Chain) WriteFreTable(outFileName string){
	outFile, err := os.Create(outFileName)
    if err != nil {
    	fmt.Println("Sorry: couldn’t create the file!")
	}
	defer outFile.Close()

	fmt.Fprintln(outFile, c.prefixLen)//first line is prefixLen

	for i, suffix := range c.chain{//for each prefix
		ss := strings.Split(i, " ")//Be careful: this nou work with string with spcace
		flag := false
		count := 0
		for j := 0; j < len(ss); j++{
			if len(ss[j]) == 0{ //white space goes with ""
				count++
				fmt.Fprint(outFile, "\"\"", " ")

			}else if flag != true{
				i = i[count:]
				fmt.Fprint(outFile, i, " ")
				flag = true
			}
		}
		for _, val := range suffix{//for each suffix
			fmt.Fprint(outFile, val.word, " ", val.frequency, " ")
		}
		fmt.Fprintln(outFile)
	}
}	
/*
 * ReadFreTable reads the given model file and initilize a chain.
 * The first line of model file gives prefixLen.
 * The rest, Each line of model file in format prefix Suffix{word frequency}
 */
func ReadFreTable(modelFile string) *Chain {
	in, err := os.Open(modelFile)
	if err != nil {
		fmt.Println("Sorry: couldn’t open the file")
		os.Exit(3)
	}
	defer in.Close()
	scanner := bufio.NewScanner(in)

	var prefixLen int = 0
	if(scanner.Scan()){
		prefixLen, _ = strconv.Atoi(scanner.Text())//get prefixLen
	}
	c := NewChain(prefixLen)//a new chain

	for scanner.Scan(){
		var line string
		var words []string = make([]string, 0)
		var key string
		line = scanner.Text()//get a whole line each time we scan
		words = strings.Split(line, " ")//split the line by white space
		for i := 0 ; i < prefixLen; i++{//get key of the map, which is prefix 
			key += words[i]
			key += " "
		}
		key = key[0:len(key)-1]//the last space should be eliminated as a key(prefix) of map
		for i := prefixLen; i < len(words)-1; i += 2{//get all suffix of current prefix
			var newSuf Suffix
			newSuf.word = words[i]
			newSuf.frequency, _ = strconv.Atoi(words[i+1])
			c.chain[key] = append(c.chain[key], newSuf)
		}
	}
	return c
}


//Generate returns a string of at most n words generated from Chain.
func (c *Chain) Generate(n int) string {
	p := make(Prefix, c.prefixLen)
	for i := 0; i < c.prefixLen; i++{
		p[i] = "\"\""
	}
	var words []string
	for i := 0; i < n; i++ {
		temp := p.String()
		choices := c.chain[temp]//get slices of suffix
		if len(choices) == 0 {//nothing could be generated as no key in map
			break
		}
		var sum []int = make([]int, 1000)
		var count int = 0
		//for prorportion calculation
		for j,val := range
		 choices{
			if j == 0{
				sum[j] = val.frequency
			}else{
				sum[j] = sum[j-1] + val.frequency
			}
		}
		//random num to choose, by proportion/possibility
		random := rand.Intn(sum[len(choices)-1])
		for i := 0; i < len(choices); i++{
			if random >= sum[i]{
				count++
			}
		}
		next := choices[count].word
		words = append(words, next)
		
		p.Shift(next)
	}
	return strings.Join(words, " ")
}

func main() {

	rand.Seed(time.Now().UnixNano()) // Seed the random number generator.
	
	cmd := os.Args[1]
	if cmd == "read"{
		outputFile := os.Args[3]
		num, err := strconv.Atoi(os.Args[2])
		if err != nil || num <= 0 {
			fmt.Println("Sorry: number of prefix should be positive.")
		return
		}
		var inputFile []string//inputfile into a slice
		for i := 4; i < len(os.Args); i++{
			inputFile = append(inputFile, os.Args[i])
		}
		
		c := NewChain(num)//initialize a new Chain with given prefix length
		c.Build(inputFile)//build chain with given input files 
		c.WriteFreTable(outputFile)//write chain to the output file

	}else if cmd == "generate" {
		if len(os.Args) == 4{
			model := os.Args[2]
			n, err := strconv.Atoi(os.Args[3])
			if err != nil || n <= 0 {
				fmt.Println("Sorry: number of words should be positive.")
				return
			}
			c := ReadFreTable(model)//read from model file to initialize a chain
			text := c.Generate(n)//use the chain to generate n words
			fmt.Println(text)

		}else{
			fmt.Println("Sorry: generate option needs 4 parameters in total.")
		}
	}else{
		fmt.Println("Sorry: choose read or generate for command option for 1st parameter.")
	}
}
