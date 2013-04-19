#!/usr/bin/env python
# encoding: utf-8
"""
docker_connect.py

Created by Mat Appelman on 2013-04-18.
Copyright (c) 2013 __MyCompanyName__. All rights reserved.
"""
import os
import sys
import socket

def connect(msg):
	try:
		try:
			sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
		except socket.error, msg:
			sys.stderr.write("[ERROR] %s\n" % msg[1])
			sys.exit(1)
		
		try:
			sock.connect(("ec2-23-20-84-18.compute-1.amazonaws.com", 7000))
		except socket.error, msg:
			sys.stderr.write("[ERROR] %s\n" % msg[1])
			sys.exit(2)
		
		sock.send(msg)
		
		data = sock.recv(1024)
		
		while data:
			print data
			data = sock.recv(1024)
			if not data or "closed connection" in data:
				break
			
		print "closing socket"
		sock.close()
		print "socket closed"
	except KeyboardInterrupt:
		pass
 	return

if __name__=="__main__":
	while 1:
		msg = ''
		try:
			msg = raw_input('> ')
			connect(msg)
		except:
			raise
	
	sys.exit(0)