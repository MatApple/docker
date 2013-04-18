#!/usr/bin/env python
# encoding: utf-8
"""
docker_connect.py

Created by Mat Appelman on 2013-04-17.
Copyright (c) 2013 __MyCompanyName__. All rights reserved.
"""

import sys
import os

import socket
import sys
import simplejson
import binascii
import pickle
import zerorpc

def bin(x):
    if x==0:
        return '0'
    else:
        return (bin(x/2)+str(x%2)).lstrip('0') or '0'


def docker_connect():
	HOST, PORT = "ec2-23-20-84-18.compute-1.amazonaws.com", 7000
	data = "docker ps\n"
	
	# Create a socket (SOCK_STREAM means a TCP socket)
	sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
	
	# Connect to server and send data
	sock.connect((HOST, PORT))
	sock.send(data)
	
	# Receive data from the server and shut down
	received = sock.recv(1024)
	sock.close()
	
	print("Sent:     %s" % data)
	print("Received: %s" % received)
	
	return


def docker_connect_proxy():
	c = zerorpc.Client()
	c.connect("tcp://ec2-23-20-84-18.compute-1.amazonaws.com:7000")
	print c.proxy("ps")


if __name__ == '__main__':
	docker_connect_proxy()