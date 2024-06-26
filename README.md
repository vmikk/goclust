# Clustering tool for sparse matrices produced by USEARCH

## Motivation

The 32-bit version of USEARCH cannot process large distance matrices due to memory limitations. 
This can be a significant bottleneck when working with large sequence datasets. 
To overcome this limitation, we present a tool that performs clustering similarly to the `usearch -cluster_aggd`. 
Currently, only single linkage and complete linkage methods are implemented.  

## Quick start

First, use USEARCH to calculate the distance matrix for your sequences with a maximum distance cutoff:
```bash
usearch -calc_distmx seqs.fa -tabbedout mx.txt -maxdist 0.3
```

Next, perform the clustering using the `goclust` tool:
```bash
goclust --input mx.txt --output clusters.txt --cutoff 0.01 --method single
```

This command is an alternative to the USEARCH clustering command:
```bash 
usearch -cluster_aggd mx.txt -clusterout clusters.txt -id 0.99 -linkage min
```


## Description

The input for clustering is a "sparse" distance matrix 
estimated by `usearch -calc_distmx`, 
which only stores a subset of distances, 
omitting pairs with low identities as determined by the `maxdist` threshold. 
This significantly reduces the time and space required to compute 
and store a matrix for large sequence sets. 
Missing entries in the matrix are assumed to be at the maximum possible distance of 1.0.

## Installation

Download the `goclust` binary:

```bash
wget https://github.com/vmikk/goclust/releases/download/0.1/goclust
chmod +x goclust
./goclust
``` 

## Usage

The `goclust` tool is designed for clustering sequences based on a sparse distance matrix.   

Usage example:
```bash
goclust --cutoff <float> --includeequal=<bool> --method <string> --input <file> --output <file>
```

Parameters:

- `--cutoff`: This parameter specifies the distance cutoff for clustering. The value must be a floating-point number greater than 0. Clusters are formed by linking sequences that have a pairwise distance less than this cutoff. A lower cutoff value will result in a larger number of smaller clusters, while a higher cutoff may produce fewer, larger clusters.

- `--input`: The path to the input file containing pairwise distances. This file should be a "sparse" matrix generated by `usearch -calc_distmx`, where each row contains the distances between a pair of sequences.

- `--output`: The path to the output file where the cluster assignments will be saved. The output file will list each sequence along with its assigned cluster label.

- `--includeequal`: This option determines whether distances equal to the specified cutoff should be included in the clustering process. By default, this option is set to true (`--includeequal=true`), allowing sequences with pairwise distances exactly equal to the cutoff to be included in the same cluster. Setting this option to false (`--includeEqual=false`) changes the clustering to only consider pairwise distances strictly greater than the cutoff value, potentially leading to more, smaller clusters.

- `--method`: Specifies the clustering method to use. Choose `single` for single linkage where a sequence joins a cluster if it is close to any sequence within the cluster, allowing larger clusters with no upper bound on diameter. Choose `complete` for complete linkage (equivalent to maximum linkage), where all sequences in a cluster must be within a certain distance threshold from each other, resulting in generally smaller clusters. The default setting is `single`.

## Benchmarks

### Equivalency of results

Clustering results obtained with `goclust` closely match 
those obtained with `usearch -cluster_aggd`, except for the differences in cluster labels.
The [Rand index](https://en.wikipedia.org/wiki/Rand_index) between the two methods is 1, indicating perfect agreement.

### Performance benchmark

**Input data**: `mx.txt` - sparse distance matrix, 24MB, 1,468 unique sequences, 841,080 lines.  

Performance comparisons are conducted using 
`goclust` v.0.2 (ex-`single_linkage`), 
`usearch` v.11.0.667 (i86linux32), 
and `hyperfine` v.1.18.0:  

```bash
hyperfine \
  --warmup 3 --runs 5 \
  --export-markdown SING_BENCH.md \
  "usearch -cluster_aggd mx.txt -clusterout clusters_USEARCH.txt -id 0.99 -linkage min" \
  "./goclust --input mx.txt --output clusters_SL.txt --cutoff 0.01 --method single"
```

The benchmark results are as follows:

| Command                 |      Mean [s] | Min [s] | Max [s] |    Relative |
|:------------------------| -------------:| -------:| -------:| -----------:|
| `usearch -cluster_aggd` | 2.593 ± 0.467 |   2.220 |   3.160 | 5.96 ± 3.93 |
| `goclust`               | 0.435 ± 0.276 |   0.218 |   0.881 |        1.00 |


Processing of a larger file (11GB, 29,278 unique sequences, 393,645,092 lines), which `usearch -cluster_aggd` fails to handle due to memory limitations, takes approximately 144 seconds.
