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

// getSingleLinkageClusters reads pairwise distances from the input file,
// forms clusters based on the cutoff distance,
// and returns cluster members and their IDs
func getSingleLinkageClusters(inputPath string, cutOff float64, includeEqual bool) ([]clusterInfo, error) {
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

		// Comparison with the cutoff depends on the `--includeequal` flag
		if includeEqual && distance > cutOff || !includeEqual && distance >= cutOff {
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

	// Reassign cluster IDs to be zero-based and sequential
	sequentialClusters := reassignClusterIDs(clustersID)

	return sequentialClusters, nil
}

// Complete linkage clustering
func getCompleteLinkageClusters(inputPath string, cutOff float64, includeEqual bool) ([]clusterInfo, error) {
	clustersID := make(map[string]int)
	clusterMembers := make(map[int]map[string]bool)
	maxDistances := make(map[int]float64) // track maximum distances per cluster
	// labelsSet := make(map[string]bool)

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
			continue // Skip insufficient data
		}
		label1, label2, distanceStr := parts[0], parts[1], parts[2]

		distance, err := strconv.ParseFloat(distanceStr, 64)
		if err != nil {
			return nil, err
		}

		in1, ok1 := clustersID[label1]
		in2, ok2 := clustersID[label2]

		if !ok1 && !ok2 {
			// Both labels are new, create a new cluster
			clusterID := numClusters
			clustersID[label1] = clusterID
			clustersID[label2] = clusterID
			clusterMembers[clusterID] = map[string]bool{label1: true, label2: true}
			maxDistances[clusterID] = distance
			numClusters++
		} else if ok1 && !ok2 {
			clustersID[label2] = in1
			clusterMembers[in1][label2] = true
			updateMaxDistance(maxDistances, distance, in1)
		} else if !ok1 && ok2 {
			clustersID[label1] = in2
			clusterMembers[in2][label1] = true
			updateMaxDistance(maxDistances, distance, in2)
		} else if in1 != in2 {
			// Based on maxDistances, decide whether to merge or not
			shouldMerge := (includeEqual && maxDistances[in1] <= cutOff) || (!includeEqual && maxDistances[in1] < cutOff)
			if shouldMerge {
				// Merge the clusters
				mergeClusters(clusterMembers, clustersID, in1, in2)
				// After merging, merge their maxDistances and recompute
				maxDistances[in1] = max(maxDistances[in1], maxDistances[in2])
				delete(maxDistances, in2)
			}
		}
	}

	// Reassign cluster IDs to be zero-based and sequential
	sequentialClusters := reassignClusterIDs(clustersID)

	return sequentialClusters, nil
}


// Function to reassign cluster IDs sequentially
func reassignClusterIDs(clustersID map[string]int) []clusterInfo {
	newID := 0
	oldToNewID := make(map[int]int)
	var clusters []clusterInfo

	for label, oldID := range clustersID {
		if _, exists := oldToNewID[oldID]; !exists {
			oldToNewID[oldID] = newID
			newID++
		}
		clusters = append(clusters, clusterInfo{Label: label, ClusterID: oldToNewID[oldID]})
	}

	return clusters
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
	includeEqual := flag.Bool("includeequal", true, "Include distances equal to cutoff in clustering (default is true; set it to false for strictly greater than cutoff)")

	flag.Parse()

	if *input == "" || *output == "" || *cutoff == 0.0 {
		log.Println("Input, output, and cutoff parameters are required.")
		flag.Usage()
		return
	}

	clusters, err := getSingleLinkageClusters(*input, *cutoff, *includeEqual)
	if err != nil {
		log.Fatalf("Error processing clusters: %v", err)
	}

	if err := exportClusters(*output, clusters); err != nil {
		log.Fatalf("Error exporting clusters: %v", err)
	}
}
