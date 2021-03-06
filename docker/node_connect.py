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


def talk(msg,host,port):
	try:
		sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
	except socket.error, msg:
		sys.stderr.write("[ERROR] %s\n" % msg[1])
		sys.exit(1)
	try:
		sock.connect((host, port))
	except socket.error, msg:
		sys.stderr.write("[ERROR] %s\n" % msg[1])
		sys.exit(2)
	sock.send(msg)
	data = sock.recv(1024)
	while data:
		if data and len(data) > 0: 
			print data
		data = sock.recv(1024)
		if not data or "closed connection" in data:
			break
	sock.close()
	print "socket closed"
 	return



if __name__=="__main__":
	HOST = "ec2-23-20-122-37.compute-1.amazonaws.com"
	PORT = 4242
	while 1:
		msg = ''
		try:
			msg = raw_input('> ')
			talk(msg,HOST,PORT)
		except:
			raise
	
	sys.exit(0)