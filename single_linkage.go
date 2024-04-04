package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func getSingleLinkageClusters(inputPath string, cutOff float64) ([]string, []int) {
	clustersID := make(map[string]int)
	clusterMembers := make(map[int]map[string]bool)
	labelsSet := make(map[string]bool)

	numClusters := 0
	file, err := os.Open(inputPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		label1, label2, distanceStr := parts[0], parts[1], parts[2]

		if _, exists := labelsSet[label1]; !exists {
			labelsSet[label1] = true
		}

		distance, err := strconv.ParseFloat(distanceStr, 64)
		if err != nil {
			panic(err)
		}

		if distance >= cutOff {
			continue
		}

		in1, ok1 := clustersID[label1]
		in2, ok2 := clustersID[label2]

		if !ok1 && !ok2 {
			clustersID[label1] = numClusters
			clustersID[label2] = numClusters
			clusterMembers[numClusters] = make(map[string]bool)
			clusterMembers[numClusters][label1] = true
			clusterMembers[numClusters][label2] = true
			numClusters++
		} else if ok1 && !ok2 {
			clustersID[label2] = in1
			clusterMembers[in1][label2] = true
		} else if !ok1 && ok2 {
			clustersID[label1] = in2
			clusterMembers[in2][label1] = true
		} else if in1 != in2 {
			for label := range clusterMembers[in2] {
				clustersID[label] = in1
				clusterMembers[in1][label] = true
			}
			delete(clusterMembers, in2)
		}
	}

	labelsList := make([]string, 0, len(labelsSet))
	for label := range labelsSet {
		labelsList = append(labelsList, label)
	}

	clusterLabels := make([]int, len(labelsList))
	for i, label := range labelsList {
		clusterLabels[i] = clustersID[label]
	}

	return labelsList, clusterLabels
}

func exportClusters(outputPath string, labelsList []string, clusterLabels []int) {
	file, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for i, label := range labelsList {
		fmt.Fprintf(file, "%d\t%s\n", clusterLabels[i], label)
	}
}

func main() {
	input := flag.String("input", "", "Path to the input sparse matrix file")
	output := flag.String("output", "", "Path to the output file")
	cutoff := flag.Float64("cutoff", 0.0, "Distance cutoff for clustering")
	flag.Parse()

	if *input == "" || *output == "" || *cutoff == 0.0 {
		fmt.Println("Input, output, and cutoff parameters are required.")
		flag.Usage()
		os.Exit(1)
	}

	labelsList, clusterLabels := getSingleLinkageClusters(*input, *cutoff)
	exportClusters(*output, labelsList, clusterLabels)
}
