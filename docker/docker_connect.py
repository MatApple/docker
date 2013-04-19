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
from subprocess import PIPE, STDOUT, Popen, CalledProcessError
from threading  import Thread
import zerorpc

import eventlet

try:
    from Queue import Queue, Empty
except ImportError:
    from queue import Queue, Empty  # python 3.x

ON_POSIX = 'posix' in sys.builtin_module_names


def closed_callback():
    print "called back"

def enqueue_output(out, queue, proc):
	while proc.poll() is None:
		for line in iter(out.readline, ""):
			if line=="":
				break
			queue.put(line)
		eventlet.sleep(0.1)
	queue.put("closed connection")
	err.close()
	out.close()

class Docker(object):
	def __init__(self, client, cb):
		self.q=Queue()
		self.client=client
		self.cb=cb
		self.StdIn()
		
	def StdIn(self):
	    while True:
	        d = self.client.recv(32384)
	        if d == '':
	            self.cb()
	            break
	        self.runCommand(d)

	def runCommand(self,cmd):
		print "command: ",cmd
		p = Popen([cmd], stdout=PIPE, stderr=STDOUT, shell=True, close_fds=ON_POSIX)
		t = Thread(target=enqueue_output, args=(p.stdout, self.q, p))
		t.daemon = True # thread dies with the program
		t.start()
		self.stdOut()
	
	def stdOut(self):
		while True:
			try:
				msg = self.q.get_nowait() # or q.get(timeout=.1)
			except Empty:
				eventlet.sleep(0.1)
			else:
				self.client.sendall(msg)
				if msg == "closed connection":
					break
			


listener = eventlet.listen(('0.0.0.0', 7000))
print "listening!"
try:
	while True:
		client, addr = listener.accept()
		eventlet.spawn_n(Docker, client, closed_callback)
except KeyboardInterrupt:
	sys.exit(0)












"""

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
"""