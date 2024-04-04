## Author - Felipe González Casabianca
## 2024, April 03

def get_single_linkage_clusters(input_path, cut_off):
    '''
    Method that computes the single linkage clusters from the given sparse matrix file.
    Note: Cutoff is assumed to be inclusive, that is: clusters will not be merged if the distance
    is larger or equal to the given cutoff
    '''

    clusters_id = {}
    cluster_members = {}
    labels_set = set()

    num_clusters = 0
    with open(input_path, 'r') as file:
        for line in file:
            label1, label2, distance = line.strip().split()

            # Stop condition
            if len(labels_set) > 0 and len(labels_set) == len(clusters_id):
                # Done
                print("Finished")
                break

            if label1 == label2:
                labels_set.add(label1)
            else:

                distance = float(distance)

                # Checks for distance and cutoff
                if distance >= cut_off:
                    continue

                # Checks if labels already have been assigned
                in1 = label1 in clusters_id
                in2 = label2 in clusters_id
  
                # Cases
                if in1 and not in2:
                    clusters_id[label2] = clusters_id[label1]
                    cluster_members[clusters_id[label1]].add(label2)
                elif in2 and not in1:
                    clusters_id[label1] = clusters_id[label2]
                    cluster_members[clusters_id[label2]].add(label1)
                elif not in2 and not in1:
                    clusters_id[label1] = num_clusters
                    clusters_id[label2] = num_clusters
                    cluster_members[num_clusters] = set([label1, label2])
                    num_clusters += 1
                else:
                    if clusters_id[label1] == clusters_id[label2]:
                        continue
                    else:
                        # merges
                        to_be_merged = clusters_id[label1]
                        to_be_deleted = clusters_id[label2]
                        cluster_members[to_be_merged].update(cluster_members[to_be_deleted])
                        # Updates
                        for k in cluster_members[to_be_deleted]:
                            clusters_id[k] = to_be_merged

                        del cluster_members[to_be_deleted]

    # Adds isolated labels
    for lab in labels_set.difference(clusters_id.keys()):
        clusters_id[lab] = num_clusters
        cluster_members[num_clusters] = set([lab])
        num_clusters += 1


    labels_list = list(labels_set)
    
    # Cleans so cluster labels are continuos
    new_cluster_ids = {}
    new_num_clusters = 0 
    for _, value in cluster_members.items():
        new_cluster_ids.update({key: new_num_clusters for key in value})
        new_num_clusters += 1


    cluster_labels = [new_cluster_ids[k] for k in labels_list]

    return labels_list, cluster_labels
