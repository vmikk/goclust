# Single linkage clustering of sparse matrix

## Motivation

The 32-bit version of USEARCH cannot process large distance matrices due to memory limitations. 
This can be a significant bottleneck when working with large sequence datasets. 
To overcome this limitation, we present a tool that performs single linkage clustering 
similarly to the `usearch -cluster_aggd`.

## Quick start

First, use USEARCH to calculate the distance matrix for your sequences with a maximum distance cutoff:
```bash
usearch -calc_distmx seqs.fa -tabbedout mx.txt -maxdist 0.3
```

Next, perform the clustering using the single_linkage tool:
```bash
single_linkage --input mx.txt --output clusters.txt --cutoff 0.01
```

This command is an alternative to the USEARCH clustering command:
```bash 
usearch -cluster_aggd mx.txt -clusterout clusters.txt -id 0.99 -linkage min
```

