# -*- mode: ruby -*-
# vi: set ft=ruby :
#
require 'fileutils'

required_dirs = ["private"]
required_dirs.each do |dir|
  full_path = File.join(Dir.pwd, dir)
  FileUtils.mkdir_p(full_path)
end

NUM_WORKERS = 2 # Set number of worker nodes
VAGRANT_BOX = "ubuntu/jammy64"
APISERVER_ADVERTISE_ADDRESS = "172.16.56.10"

INIT_FILE="/private/kubeadm-init-config.yaml"
JOIN_FILE="/private/kubeadm-join-config"
JOIN_FILE_TEMPLATE="#{JOIN_FILE}.yaml.tmpl"

Vagrant.configure("2") do |config|
  config.vm.box = VAGRANT_BOX
  
  # Private is created when vagrantfile is ran. This is where tmp files are generated and shared 
  # between VMs. Stuff with short lived credentials, kubeconfigs, ect.
  config.vm.synced_folder "./private", "/private", create: true, mount_options: ["dmode=777", "fmode=666"]

  config.vm.define "control-plane" do |cp|
    cp.vm.hostname = "k8s-control-plane"
    cp.vm.network "private_network", ip: APISERVER_ADVERTISE_ADDRESS
    cp.vm.provider "virtualbox" do |vb|
      vb.memory = 4096
      vb.cpus = 2
    end
    cp.vm.provision "ansible" do |ansible|
      ansible.playbook = "playbooks/setup.yml"
      ansible.extra_vars = {
        node_role: "control-plane",
        apiserver_advertise_address: APISERVER_ADVERTISE_ADDRESS,
        node_ip: APISERVER_ADVERTISE_ADDRESS,
        init_file: INIT_FILE,
        join_file_template: JOIN_FILE_TEMPLATE
      }
    end
  end

  (1..NUM_WORKERS).each do |i|
    config.vm.define "worker-#{i}" do |worker|
      HOSTNAME = "k8s-worker-#{i}"
      NODE_IP="172.16.56.#{10 + i}"
      worker.vm.hostname = HOSTNAME
      worker.vm.network "private_network", ip: NODE_IP
      worker.vm.provider "virtualbox" do |vb|
        vb.memory = 2048
        vb.cpus = 2
      end
      worker.vm.provision "ansible" do |ansible|
        ansible.playbook = "playbooks/setup.yml"
        ansible.extra_vars = { 
          node_role: "worker",
          node_ip: NODE_IP,
          join_file_template: JOIN_FILE_TEMPLATE,
          node_join_file: "#{JOIN_FILE}_#{HOSTNAME}.yaml"
        }
      end
    end
  end
end
