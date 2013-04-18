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
	
	@zerorpc.stream
	def proxy(self,data): 
		c = zerorpc.Client()
		c.connect("tcp://127.0.0.1:4242")
		return c



try:
	s = zerorpc.Server(DockerConnect())
	s.bind("tcp://0.0.0.0:7000")
	s.run()
except KeyboardInterrupt:
	sys.exit(0)


