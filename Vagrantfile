# -*- mode: ruby -*-
# vi: set ft=ruby :

def v10(config)
  config.vm.box = "quantal64_3.5.0-25"
  config.vm.box_url = "http://get.docker.io/vbox/ubuntu/12.10/quantal64_3.5.0-25.box"

  config.vm.share_folder "v-data", "/opt/go/src/github.com/dotcloud/docker", File.dirname(__FILE__)

  # Ensure puppet is installed on the instance
  config.vm.provision :shell, :inline => "apt-get -qq update; apt-get install -y puppet; sudo apt-get install linux-image-extra-`uname -r`;"

  config.vm.provision :puppet do |puppet|
    puppet.manifests_path = "puppet/manifests"
    puppet.manifest_file  = "quantal64.pp"
    puppet.module_path = "puppet/modules"
  end
end

Vagrant::VERSION < "1.1.0" and Vagrant::Config.run do |config|
  v10(config)
end

Vagrant::VERSION >= "1.1.0" and Vagrant.configure("1") do |config|
  v10(config)
end

Vagrant::VERSION >= "1.1.0" and Vagrant.configure("2") do |config|
  config.vm.provider :aws do |aws, override|
    config.vm.box = "dummy"
    config.vm.box_url = "https://github.com/mitchellh/vagrant-aws/raw/master/dummy.box"
    aws.access_key_id = "AKIAI7CYRNSO56TF3SBQ"
    aws.secret_access_key = "PGvhPNW/HIKe3MCNnDPsm8ARVuYZfdMvbfWdnlfS"
    aws.keypair_name = "protobox"
	override.ssh.username = "ubuntu"
    override.ssh.private_key_path = "/Users/matappelman/aws/protobox.pem"
    aws.region = "us-east-1"
    aws.ami = "ami-3fec7956"
	aws.security_groups = ["protobox_host"]
    aws.instance_type = "t1.micro"
  end

  config.vm.provider :rackspace do |rs|
    config.vm.box = "dummy"
    config.vm.box_url = "https://github.com/mitchellh/vagrant-rackspace/raw/master/dummy.box"
    config.ssh.private_key_path = ENV["RS_PRIVATE_KEY"]
    rs.username = ENV["RS_USERNAME"]
    rs.api_key  = ENV["RS_API_KEY"]
    rs.public_key_path = ENV["RS_PUBLIC_KEY"]
    rs.flavor   = /512MB/
    rs.image    = /Ubuntu/
  end

  config.vm.provider :virtualbox do |vb|
    config.vm.box = "quantal64_3.5.0-25"
    config.vm.box_url = "http://get.docker.io/vbox/ubuntu/12.10/quantal64_3.5.0-25.box"
  end
end
