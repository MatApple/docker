#!/usr/bin/env python
# encoding: utf-8
"""

Created by Mat Appelman on 2013-04-17.
Copyright (c) 2013 __MyCompanyName__. All rights reserved.
"""

import sys
import os

import socket

import zerorpc



class DockerConnect(object):
	
	def proxy(self,data):
		# Create a socket (SOCK_STREAM means a TCP socket)
		sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
		
		# Connect to server and send data
		sock.connect(("127.0.0.1", 4242))
		sock.send(data)
		
		# Receive data from the server and shut down
		received = sock.recv(1024)
		sock.close()
		
		print("Sent:     %s" % data)
		print("Received: %s" % received)
		return received



try:
	s = zerorpc.Server(DockerConnect())
	s.bind("tcp://0.0.0.0:7000")
	s.run()
except KeyboardInterrupt:
	sys.exit(0)


