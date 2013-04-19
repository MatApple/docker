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

import eventlet

ON_POSIX = 'posix' in sys.builtin_module_names




def closed_callback():
    print "called back"



def enqueue_output(out, queue, proc, killq):
	killsig=False
	while proc.poll() is None and not killsig:
		for line in iter(out.readline, ""):
			if len(line) > 0:
				queue.put(line)
		try:
			sig = killq.get_nowait() # or q.get(timeout=.1)
		except:
			pass
		else:
			if sig=="KILL":
				killsig=True
				break
		eventlet.sleep(0.1)
	out.close()
	queue.put("closed connection")



class Docker(object):
	def __init__(self, client, cb):
		self.q=eventlet.queue.LightQueue()
		self.killq=eventlet.queue.LightQueue()
		self.client=client
		self.thread=None
		self.proc=None
		self.killed=False
		self.cb=cb
		self.StdIn()
		
	def StdIn(self):
		while True:
			d = self.client.recv(32384)
			## listen for closed client connections and hit kill
			if need_to_kill: self.kill_and_close()
			if d == '':
				self.cb()
				break
			self.runCommand(d)
		return

	def runCommand(self,cmd):
		print "command: ",cmd
		self.proc = Popen([cmd], stdout=PIPE, stderr=STDOUT, shell=True, close_fds=ON_POSIX)
		self.thread = Thread(target=enqueue_output, args=(p.stdout, self.q, self.proc, self.killq))
		self.thread.start()
		self.stdOut()
		return
	
	def stdOut(self):
		while True:
			try:
				msg = self.q.get_nowait() # or q.get(timeout=.1)
			except:
				eventlet.sleep(0.1)
			else:
				try:
					self.client.sendall(msg)
				except:
					if not self.killed: self.kill_and_close()
				if msg == "closed connection" and self.q.empty():
					break
				if msg == "closed connection" and self.killed:
					break
		return
	
	def kill_and_close(self):
		if self.proc: self.proc.kill()
		self.killq.put("KILL")
		self.killed=True
		return

			


listener = eventlet.listen(('0.0.0.0', 4243))
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