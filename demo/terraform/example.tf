provider "aws" {
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
  region     = "eu-west-2"
}

resource "aws_subnet" "babblenet" {
  vpc_id     = "${var.vpc}"
  cidr_block = "10.0.1.0/24"
  map_public_ip_on_launch="true"

  tags {
    Name = "Testnet"
  }
}

resource "aws_security_group" "babblesec" {
    name = "babblesec"
    description = "Babble internal traffic + maintenance."

    vpc_id     = "${var.vpc}"

    // These are for internal traffic
    ingress {
        from_port = 0
        to_port = 65535
        protocol = "tcp"
        self = true
    }

    ingress {
        from_port = 0
        to_port = 65535
        protocol = "udp"
        self = true
    }

    // These are for maintenance
    ingress {
        from_port = 22
        to_port = 22
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }
    
    ingress {
        from_port = 80
        to_port = 80
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }

    ingress {
        from_port = 8080
        to_port = 8080
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }

    ingress {
        from_port = 9090
        to_port = 9090
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }

    ingress {
        from_port = 1338
        to_port = 1338 
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }

    ingress {
        from_port = 1339
        to_port = 1339 
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }

    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }

    // This is for outbound internet access
    egress {
        from_port = 0
        to_port = 0
        protocol = "-1"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_instance" "server" {
  count = "${var.servers}"
  
  //custom ami with ubuntu + babble + evm-babble + node.js + 80->3000
  ami = "ami-768f9212" 
  instance_type = "t2.micro"

  subnet_id = "${aws_subnet.babblenet.id}"
  vpc_security_group_ids  = ["${aws_security_group.babblesec.id}"]
  private_ip = "10.0.1.${10+count.index}"

  key_name = "${var.key_name}"
  connection {
    user = "ubuntu"
    private_key = "${file("${var.key_path}")}"
  }

  provisioner "file" {
    source      = "conf/node${count.index +1}/babble"
    destination = "babble_conf" 
  }

  provisioner "file" {
    source      = "conf/node${count.index +1}/eth"
    destination = "eth_conf" 
  }

  provisioner "local-exec" {
    command = "mkdir -p conf/node${count.index +1}/web && ./scripts/build-web-config.sh ${count.index +1} ${self.public_ip} 8080 9090 conf/node${count.index +1}/web/config.json"
  }  

  provisioner "file" {
    source      = "../web/demo-server-light" #without node_modules
    destination = "demo-server" 
  }     
 
  provisioner "file" {
    source      = "conf/node${count.index +1}/web/config.json"
    destination = "demo-server/config/config.json"
  }
 
 provisioner "remote-exec" {
    inline = ["cd demo-server && npm install"]
  }

  provisioner "local-exec" {
    command = "echo ${self.private_ip} ${self.public_ip}  >> ips.dat"
  }

  #Instance tags
  tags {
      Name = "node${count.index}"
  }
}

output "public_addresses" {
    value = ["${aws_instance.server.*.public_ip}"]
}
