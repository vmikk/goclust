package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

// clusterInfo holds information about a single cluster member
type clusterInfo struct {
	Label     string
	ClusterID int
}

// getSingleLinkageClusters reads pairwise distances from the input file, forms clusters based on the cutoff distance, and returns cluster members and their IDs
func getSingleLinkageClusters(inputPath string, cutOff float64) ([]clusterInfo, error) {
	clustersID := make(map[string]int)
	clusterMembers := make(map[int]map[string]bool)
	labelsSet := make(map[string]bool)

	numClusters := 0
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue // Skip lines that don't have enough parts
		}
		label1, label2, distanceStr := parts[0], parts[1], parts[2]

		labelsSet[label1] = true
		labelsSet[label2] = true

		distance, err := strconv.ParseFloat(distanceStr, 64)
		if err != nil {
			return nil, err
		}

		if distance >= cutOff {
			continue
		}

		in1, ok1 := clustersID[label1]
		in2, ok2 := clustersID[label2]

		if !ok1 && !ok2 {
			clusterID := numClusters
			clustersID[label1] = clusterID
			clustersID[label2] = clusterID
			clusterMembers[clusterID] = map[string]bool{label1: true, label2: true}
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

	var clusters []clusterInfo
	for label, id := range clustersID {
		clusters = append(clusters, clusterInfo{Label: label, ClusterID: id})
	}

	return clusters, nil
}

// exportClusters writes the cluster members and their IDs to the output file, sorted first by cluster ID and then by label
func exportClusters(outputPath string, clusters []clusterInfo) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Sort clusters by ClusterID, then by Label
	sort.Slice(clusters, func(i, j int) bool {
		if clusters[i].ClusterID == clusters[j].ClusterID {
			return clusters[i].Label < clusters[j].Label
		}
		return clusters[i].ClusterID < clusters[j].ClusterID
	})

	for _, cluster := range clusters {
		if _, err := fmt.Fprintf(file, "%d\t%s\n", cluster.ClusterID, cluster.Label); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	input := flag.String("input", "", "Path to the input file containing pairwise distances")
	output := flag.String("output", "", "Path to the output file for cluster assignments")
	cutoff := flag.Float64("cutoff", 0.0, "Distance cutoff for clustering (must be greater than 0)")
	flag.Parse()

	if *input == "" || *output == "" || *cutoff == 0.0 {
		log.Println("Input, output, and cutoff parameters are required.")
		flag.Usage()
		return
	}

	clusters, err := getSingleLinkageClusters(*input, *cutoff)
	if err != nil {
		log.Fatalf("Error processing clusters: %v", err)
	}

	if err := exportClusters(*output, clusters); err != nil {
		log.Fatalf("Error exporting clusters: %v", err)
	}
}
