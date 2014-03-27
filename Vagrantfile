# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "precise64"
  config.vm.provider :vmware_fusion do |v|
    v.vmx['memsize'] = 256
  end

  script = <<SCRIPT
apt-get install -y tinyproxy curl iptables telnet tcpdump
SCRIPT

  config.vm.provision "shell", inline: script
end
