# Intro
This repo is a playground inspired by the NATless IPv4/IPv6 translation (presented here https://www.youtube.com/watch?v=E-X8LoLl0CM). The Linux kernel never ceases to surprise me in what its capable of. When I discovered you could swap FDs within a connect syscall, I had to try it myself. 

# Impl
This repo uses virtualbox and vagrant to provision a k8s clusters. Probably this would be achievable with KinD, though I was not confident that KinD could handle the kernel/host/pod boundray well. Sometimes KinD is a bit finiky when kernel things are involved, though this may not be so bad anymore. I have reused this vagrant/virtualbox config for a few other things so it was no issue for me to reuse it.

# Why Do Stuff Like This In General?
There seems to be broad push in container networking in general to push the limits of the host/container boundary such that container networking is as good as (or better :D) than networking on the host iteslf. So the progress was

1) Get container networking to work - Use Linux network namespaces in a clever way to segment workloads on a host such that the workloads spin up fast, are pacakged, and can be managed by an orchestrator
2) Improve/Mature container networking - IPVLAN, netfilter, ect. Pod to pod networking traverses boundaries but the work is done efficiently
3) Extend container networking for container based use cases (eBPF, netkit, cilium type stuff) - Pod to pod networking cares less about boundaries as use cases are more defined. Also eBPF gives ability to build k8s workloads in the kernel.

# Why Do NATless?
Per packet processing can be expensive and cloud providers sometimes make you pay for things that solve this problem.
