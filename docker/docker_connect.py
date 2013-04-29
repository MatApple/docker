#!/usr/bin/env python
# encoding: utf-8
"""
docker_connect.py

Created by Mat Appelman on 2013-04-18.
Copyright (c) 2013 __MyCompanyName__. All rights reserved.



import sys
import subprocess
import random
import time
import threading
import Queue



ON_POSIX = 'posix' in sys.builtin_module_names


class AsynchronousFileReader(threading.Thread):
    '''
    Helper class to implement asynchronous reading of a file
    in a separate thread. Pushes read lines on a queue to
    be consumed in another thread.
    '''
 
    def __init__(self, fd, queue):
        assert isinstance(queue, Queue.Queue)
        assert callable(fd.readline)
        threading.Thread.__init__(self)
        self._fd = fd
        self._queue = queue
 
    def run(self):
        '''The body of the tread: read lines and put them on the queue.'''
        for line in iter(self._fd.readline, ''):
            self._queue.put(line)
 
    def eof(self):
        '''Check whether there is no more content to expect.'''
        return not self.is_alive() and self._queue.empty()
 
def consume(command):
    '''
    Example of how to consume standard output and standard error of
    a subprocess asynchronously without risk on deadlocking.
    '''
 
    # Launch the command as subprocess.
    process = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=False, close_fds=ON_POSIX)
 
    # Launch the asynchronous readers of the process' stdout and stderr.
    stdout_queue = Queue.Queue()
    stdout_reader = AsynchronousFileReader(process.stdout, stdout_queue)
    stdout_reader.start()
    stderr_queue = Queue.Queue()
    stderr_reader = AsynchronousFileReader(process.stderr, stderr_queue)
    stderr_reader.start()
 
    # Check the queues if we received some output (until there is nothing more to get).
    while not stdout_reader.eof() or not stderr_reader.eof():
        # Show what we received from standard output.
        while not stdout_queue.empty():
            line = stdout_queue.get()
            print 'Received line on standard output: ' + repr(line)
 
        # Show what we received from standard error.
        while not stderr_queue.empty():
            line = stderr_queue.get()
            print 'Received line on standard error: ' + repr(line)
 
        # Sleep a bit before asking the readers again.
        time.sleep(.1)
 
    # Let's be tidy and join the threads we've started.
    stdout_reader.join()
    stderr_reader.join()
 
    # Close subprocess' file descriptors.
    process.stdout.close()
    process.stderr.close()
 

 
if __name__ == '__main__':
    # The main flow:
    command = sys.args()
    if not command:
      command = ['docker', "ps", sys.argv[0]]
    consume(command)


"""

import sys
import os
import unittest
import sys
import sys
from subprocess import PIPE, STDOUT, Popen, CalledProcessError
from threading  import Thread

import eventlet





def closed_callback():
    print "called back"



def enqueue_output(out, queue, proc):
	while proc.poll() is None:
		for line in iter(out.readline, ""):
			queue.put(line)
		eventlet.sleep(0.1)
	out.close()
	queue.put("closed connection")



class Docker(object):
	def __init__(self, client, cb):
		self.q=eventlet.queue.LightQueue()
		self.client=client
		self.proc=None
		self.killed=False
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
		self.proc = Popen([cmd], stdout=PIPE, stderr=STDOUT, shell=True, close_fds=ON_POSIX)
		t = Thread(target=enqueue_output, args=(self.proc.stdout, self.q, self.proc))
		t.daemon=True
		t.start()
		self.stdOut()

	
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
					if not self.killed: 
						self.kill_and_close()
				if msg == "closed connection" and self.q.empty():
					break
				if msg == "closed connection" and self.killed:
					break


	
	def kill_and_close(self):
		print "Killing remote client connection"
		if self.proc: 
			self.proc.kill()
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