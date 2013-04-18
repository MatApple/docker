#!/usr/bin/env python
# encoding: utf-8
"""
docker_connect.py

Created by Mat Appelman on 2013-04-18.
Copyright (c) 2013 __MyCompanyName__. All rights reserved.
"""

import sys
import os
import unittest
import sys
import sys
from subprocess import PIPE, Popen
from threading  import Thread
import zerorpc
#import gevent
#import eventlet

try:
    from Queue import Queue, Empty
except ImportError:
    from queue import Queue, Empty  # python 3.x

ON_POSIX = 'posix' in sys.builtin_module_names

def enqueue_output(self, out, queue):
	for line in iter(out.readline, b''):
		queue.put(line)
	out.close()

class Docker(object):
	def __init__(self):
		self.queue=Queue()
	
	def subscribe(self):
		try:
			for msg in self.queue:
				yield msg
		finally:
			pass
	
	@zerorpc.stream
	def runCommand(self,cmd):
		p = Popen([cmd], stdout=PIPE, bufsize=1, close_fds=ON_POSIX)
		t = Thread(target=enqueue_output, args=(p.stdout, self.queue))
		t.daemon = True # thread dies with the program
		t.start()
		self.queue.put("running cmd: "+str(cmd))
		self.subscribe()

s = zerorpc.Server(Docker())
s.bind("tcp://"+str(socket.gethostname())+":5000")
s.run()