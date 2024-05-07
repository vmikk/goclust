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

// clusterInfo holds information about a single cluster
type clusterInfo struct {
	ID      int      // ID of the cluster
	Members []string // Slice of labels representing members of the cluster
}

// There are two clustering functions - `getSingleLinkageClusters` and `getCompleteLinkageClusters`
// They read pairwise distances from the input file,
// form clusters based on the cutoff distance,
// and return cluster members and their IDs

// Single linkage clustering
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

	// Build initial clusterInfo slice from clusterMembers
    initialClusters := buildClusterInfo(clusterMembers, clustersID)

	// Reassign cluster IDs to be zero-based and sequential
	sequentialClusters := reassignClusterIDs(initialClusters)

	return sequentialClusters, nil
}

// Complete linkage clustering
func getCompleteLinkageClusters(inputPath string, cutOff float64, includeEqual bool) ([]clusterInfo, error) {
	clustersID := make(map[string]int)
	clusterMembers := make(map[int]map[string]bool)
	maxDistances := make(map[int]map[int]float64) // track maximum distances between clusters

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

		clusterID1, ok1 := clustersID[label1]
		clusterID2, ok2 := clustersID[label2]

		if !ok1 && !ok2 {
			// Both labels are new, create a new cluster
			clusterID := numClusters
			clustersID[label1] = clusterID
			clustersID[label2] = clusterID
			clusterMembers[clusterID] = map[string]bool{label1: true, label2: true}
			numClusters++
			maxDistances[clusterID] = make(map[int]float64)
		} else if ok1 && !ok2 {
			clustersID[label2] = clusterID1
			clusterMembers[clusterID1][label2] = true
			updateMaxDistancesForNewMember(clusterID1, label2, distance, clusterMembers, maxDistances, cutOff)
		} else if !ok1 && ok2 {
			clustersID[label1] = clusterID2
			clusterMembers[clusterID2][label1] = true
			updateMaxDistancesForNewMember(clusterID2, label1, distance, clusterMembers, maxDistances, cutOff)
		} else if clusterID1 != clusterID2 && distance <= cutOff {
			if shouldMerge(clusterID1, clusterID2, maxDistances, cutOff) {
				mergeClusters(clusterMembers, clustersID, clusterID1, clusterID2, maxDistances)
			}
		}
	}

	clusters := buildClusterInfo(clusterMembers, clustersID)
	return reassignClusterIDs(clusters), nil
}

// Helper function to determine if two clusters should merge based on the max distances recorded
func shouldMerge(clusterID1, clusterID2 int, maxDistances map[int]map[int]float64, cutOff float64) bool {
	// Check the maximum recorded distance between these two clusters
	if maxDistance, exists := maxDistances[clusterID1][clusterID2]; exists {
		return maxDistance <= cutOff
	}
	return false
}

// Update maximum distances when a new member is added to a cluster
func updateMaxDistancesForNewMember(clusterID int, newLabel string, distance float64, clusterMembers map[int]map[string]bool, maxDistances map[int]map[int]float64, cutOff float64) {
    for otherClusterID := range clusterMembers {
        if otherClusterID != clusterID {
            // Check if there is already a recorded distance between these clusters
            if otherDistance, exists := maxDistances[clusterID][otherClusterID]; exists {
                maxDistances[clusterID][otherClusterID] = max(otherDistance, distance)
            } else {
                maxDistances[clusterID][otherClusterID] = distance
            }
        }
    }
}


// Merge two clusters into one
func mergeClusters(clusterMembers map[int]map[string]bool, clustersID map[string]int, id1, id2 int, maxDistances map[int]map[int]float64) {
	// Transfer all members from cluster id2 to id1
	for label := range clusterMembers[id2] {
		clusterMembers[id1][label] = true
		clustersID[label] = id1
	}

	// Update maxDistances for the new cluster
	// Assuming maxDistances[id1] and maxDistances[id2] are already populated
	for k, dist := range maxDistances[id2] {
		if existingDist, exists := maxDistances[id1][k]; exists {
			maxDistances[id1][k] = max(existingDist, dist)
		} else {
			maxDistances[id1][k] = dist
		}
	}

	// Remove all references to the merged cluster id2
	delete(clusterMembers, id2)
	delete(maxDistances, id2)
	for _, distMap := range maxDistances {
		delete(distMap, id2)
	}
}

// Helper function to find the maximum of two float64 values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Construct the output structure `[]clusterInfo` from the internal clustering data
func buildClusterInfo(clusterMembers map[int]map[string]bool, clustersID map[string]int) []clusterInfo {
    var clusters []clusterInfo
    for id, members := range clusterMembers {
        memberList := make([]string, 0, len(members))
        for member := range members {
            memberList = append(memberList, member)
        }
        clusters = append(clusters, clusterInfo{ID: id, Members: memberList})
    }
    return clusters
}


// Function to reassign cluster IDs sequentially
func reassignClusterIDs(clusters []clusterInfo) []clusterInfo {
    newID := 0
    oldToNewID := make(map[int]int)
    var newClusters []clusterInfo

    // We create a new list to maintain the order and reassignment
    for _, cluster := range clusters {
        if _, exists := oldToNewID[cluster.ID]; !exists {
            oldToNewID[cluster.ID] = newID
            newID++
        }
        // Copy the cluster with a new ID
        newCluster := clusterInfo{
            ID:      oldToNewID[cluster.ID],
            Members: cluster.Members,
        }
        newClusters = append(newClusters, newCluster)
    }

    return newClusters
}


// Write the cluster members and their IDs to the output file, sorted first by cluster ID and then by label
func exportClusters(outputPath string, clusters []clusterInfo) error {
    file, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer file.Close()

    // Sort clusters by ClusterID, then lexicographically by the first Label if necessary
    sort.Slice(clusters, func(i, j int) bool {
        if clusters[i].ID == clusters[j].ID {
            // Compare first member if both clusters are non-empty for lexicographical ordering
            if len(clusters[i].Members) > 0 && len(clusters[j].Members) > 0 {
                return clusters[i].Members[0] < clusters[j].Members[0]
            }
            // Fallback if one cluster is empty, the non-empty one is 'less'
            return len(clusters[i].Members) != 0
        }
        return clusters[i].ID < clusters[j].ID
    })

    for _, cluster := range clusters {
        // Join all members into a single string separated by commas
        membersStr := strings.Join(cluster.Members, ", ")
        if _, err := fmt.Fprintf(file, "%d\t%s\n", cluster.ID, membersStr); err != nil {
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
	method := flag.String("method", "single", "Clustering method to use ('single' or 'complete')")

	// Parse the command-line flags
	flag.Parse()

	if *input == "" || *output == "" || *cutoff == 0.0 {
		log.Println("Input, output, and cutoff parameters are required.")
		flag.Usage()
		return
	}

	// fmt.Printf("Using the %s method for clustering.\n", *method)

	var clusters []clusterInfo
	var err error

	if *method == "single" {
		clusters, err = getSingleLinkageClusters(*input, *cutoff, *includeEqual)
	} else {
		clusters, err = getCompleteLinkageClusters(*input, *cutoff, *includeEqual)
	}

	if err != nil {
		log.Fatalf("Error processing clusters: %v", err)
	}

	if err := exportClusters(*output, clusters); err != nil {
		log.Fatalf("Error exporting clusters: %v", err)
	}
}
