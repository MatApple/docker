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

import eventlet

def closed_callback():
    print "called back"

def forward(source, dest, cb = lambda: None):
    """Forwards bytes unidirectionally from source to dest"""
    while True:
        d = source.recv(32384)
        if d == '':
            cb()
            break
        dest.sendall(d)

listener = eventlet.listen((socket.gethostname(), 7000))

while True:
    client, addr = listener.accept()
    server = eventlet.connect(('127.0.0.1', 4242))
    # two unidirectional forwarders make a bidirectional one
    eventlet.spawn_n(forward, client, server, closed_callback)
    eventlet.spawn_n(forward, server, client)

