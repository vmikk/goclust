#!/usr/bin/python
"""
Complete linkage clustering

Author - Felipe González Casabianca
2024, May 09
"""

import sys

# Constants
C1 = "CLUSTER_1"
C2 = "CLUSTER_2"
D = "DISTANCE"


def complete_linkage(input_path, cut_off, strict = False):

    # Support Functions
    def get_distance(set1, set2, distances):

        d = 0   
        for l1 in set1:
            for l2 in set2:
                if (l1,l2) in distances:
                    d = max(d,distances[(l1,l2)])
                elif (l2,l1) in distances:
                    d = max(d,distances[(l2,l1)])
                else:
                    return -1

        return d
        

    cluster_distances = {}
    object_distances = {}
    neighbors = {}
    clusters = {}
    objects = set()


    with open(input_path, 'r') as file:
        for line in file:
            label1, label2, distance = line.strip().split()

            # Adds objects
            objects.add(label1)
            objects.add(label2)

            distance = float(distance)

            # Checks for distance and cutoff
            if label1 == label2:
                continue
            if (distance > cut_off) or (distance >= cut_off and strict): 
                continue
            
            # Alphabetical order
            if label1 > label2:
                label1, label2 = label2, label1
            
            ob = {C1 : label1, C2 : label2, D : distance}

            cluster_distances[(label1,label2)] = ob

            object_distances[(label1,label2)] = distance

            if label1 not in neighbors:
                neighbors[label1] = set()
            neighbors[label1].add(label2)

            if label2 not in neighbors:
                neighbors[label2] = set()
            neighbors[label2].add(label1)


    # Creates Individual clusters
    for ob in objects:
        clusters[ob] = [ob]

    # Starts
    cluster_distances_list = cluster_distances.values()


    while len(cluster_distances_list) > 0:

        min_element = min(cluster_distances_list, key=lambda x: x[D])

        # Extracts
        c1_label = min_element[C1]
        c2_label = min_element[C2]

        c1_objects = clusters[c1_label]
        c2_objects = clusters[c2_label]


        # Merges
        # New name is c1
        new_cluster = c1_objects + c2_objects
        clusters[c1_label] = c1_objects + c2_objects
        
        # Deletes old
        del clusters[c2_label]
        del cluster_distances[(c1_label,c2_label)]
        neighbors[c1_label].remove(c2_label)
        neighbors[c2_label].remove(c1_label)


        # Updates
        to_remove = [(c1_label, other) for other in neighbors[c1_label]]
        to_remove += [(c2_label, other) for other in neighbors[c2_label]]
        to_remove = [(l1,l2) if l1<=l2 else (l2,l1) for l1,l2 in to_remove]

        del neighbors[c2_label]

        for uc1_label, uc2_label in to_remove:
            
            if uc1_label == c1_label or uc2_label == c1_label:
                
                # Sorts labels
                other_label = uc2_label if uc1_label == c1_label else uc1_label
                other_cluster = clusters[other_label]                
                # Gets Max distance
                new_dist = get_distance(new_cluster, other_cluster, object_distances)


                if new_dist != -1:
                    cluster_distances[(uc1_label,uc2_label)] = {C1 : uc1_label, C2 : uc2_label, D : new_dist}
                else:
                    del cluster_distances[(uc1_label,uc2_label)]
                    neighbors[uc1_label].remove(uc2_label)
                    neighbors[uc2_label].remove(uc1_label)


                
            elif uc1_label == c2_label or uc2_label == c2_label:
                
                # Sorts labels
                other_label = uc2_label if uc1_label == c2_label else uc1_label
                other_cluster = clusters[other_label]    
                # Gets Max distance
                new_dist = get_distance(new_cluster, other_cluster, object_distances)

                # Removes
                neighbors[other_label].remove(c2_label)

                del cluster_distances[(uc1_label,uc2_label)]
                

                if new_dist != -1:
                    
                    # Updates Name
                    neighbors[other_label].add(c1_label)
                    neighbors[c1_label].add(other_label)

                    
                    l1, l2 = other_label, c1_label
                    if l1 > l2:
                        l1, l2 = l2, l1

                    cluster_distances[(l1,l2)] = {C1 : l1, C2 : l2, D : new_dist}



        cluster_distances_list = cluster_distances.values()

    # Return Object
    label_list = []
    membership = []

    for i, cluster in enumerate(clusters.values(), start=1):
        label_list += cluster
        membership += [i]*len(cluster)

    return label_list, membership


def main():
    if len(sys.argv) != 4:
        print("Usage: python complete_linkage.py <input_file> <output_file> <cutoff>")
        sys.exit(1)

    input_file = sys.argv[1]
    output_file = sys.argv[2]
    cutoff = float(sys.argv[3])

    labels, memberships = complete_linkage(input_file, cutoff)

    with open(output_file, 'w') as file:
        for label, membership in zip(labels, memberships):
            file.write(f"{label} {membership}\n")

if __name__ == "__main__":
    main()
