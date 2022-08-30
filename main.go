package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"strconv"
)

type lcg struct {
	multiplier int64
	addend     int64
	mask       int64
	seed       int64
}

func NewLCG() lcg {
	newLCG := lcg{}
	newLCG.multiplier = 0x5DEECE66D
	newLCG.addend = 0xB
	newLCG.mask = ((1 << 48) - 1)
	return newLCG
}

func (l *lcg) SetInternalSeed(newseed int64) {
	l.seed = newseed
	return
}

// only use this if the bound is a power of two
func (l *lcg) NextIntPow2(bound int64) int32 {
	l.seed = (l.seed*l.multiplier + l.addend) & l.mask
	// 17 = 48-31
	return int32((bound * (l.seed >> 17)) >> 31)
}

const maxUint31 = (1 << 31) - 1
const maxUint17 = (1 << 17) - 1

func CrackIt(bound int64, samples []int32, numToGen int, continueOnMatch bool) {
	firstSample := samples[0]
	followUpSamples := samples[1:]

	l := NewLCG()

	// Only the 31 left-most bits of the internal state are used to generate an output.
	// So, we can find these bits through brute-force quite easily
	for firstBits := int64(0); firstBits <= maxUint31; firstBits++ {
		if int32((bound*int64(firstBits))>>31) == firstSample {
			// We have a candidate for the 31 left-most bits
			// Try and find the final 17 bits that generate a valid state for the rest of our samples
			for i := int64(0); i <= maxUint17; i++ {
				stateGuess := firstBits<<17 + i
				l.SetInternalSeed(stateGuess)
				success := true
				for j := range followUpSamples {
					genned := l.NextIntPow2(bound)
					if genned != followUpSamples[j] {
						success = false
						break
					}
				}
				if success {
					fmt.Println("======")
					fmt.Printf("Success! Found starting state %d (%d<<17 + %d)\n", stateGuess, firstBits, i)
					fmt.Printf("Generating next %d outputs:\n\n", numToGen)
					for j := 0; j < numToGen; j++ {
						fmt.Printf("%d: %d\n", j, l.NextIntPow2(bound))
					}
					fmt.Println("======")

					if !continueOnMatch {
						return
					}
				}
			}
		}
	}
}

func main() {
	boundPtr := flag.Int("bound", 0, "Bound value (argument passed to nextInt())")
	continueOnMatchPtr := flag.Bool("continue", false, "Find all possible matches (only do this if the output was wrong the first time)")
	samplesPtr := flag.String("samples", "", "List of known outputs (comma or space separated)")
	genCountPtr := flag.Int("gen", 5, "How many values to predict")

	flag.Parse()

	samplesStrs := regexp.MustCompile("[0-9]+").FindAllString(*samplesPtr, -1)
	samples := make([]int32, len(samplesStrs))

	for i := range samplesStrs {
		x, err := strconv.Atoi(samplesStrs[i])
		if err != nil {
			log.Fatalf("Couldn't convert %s into a number: %v", samplesStrs[i], err)
		}
		samples[i] = int32(x)
	}

	bound := int64(*boundPtr)

	if bound == 0 {
		fmt.Printf("Please pass a bound\n\nUsage:\n")
		flag.PrintDefaults()
		return
	}

	if bound <= 0 || (bound&(bound-1)) != 0 {
		fmt.Printf("Invalid bound (%d) (make sure it is a positive power of 2)\n\nUsage:\n", bound)
		flag.PrintDefaults()
		return
	}

	if len(samples) < 2 {
		fmt.Printf("Please pass at least 2 samples\n\nUsage:\n")
		flag.PrintDefaults()
		return
	}

	CrackIt(bound, samples, *genCountPtr, *continueOnMatchPtr)
}
