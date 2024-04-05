# Single linkage clustering of sparse matrix

## Motivation

The 32-bit version of USEARCH cannot process large distance matrices due to memory limitations. 
This can be a significant bottleneck when working with large sequence datasets. 
To overcome this limitation, we present a tool that performs single linkage clustering 
similarly to the `usearch -cluster_aggd`.
