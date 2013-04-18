#!/usr/bin/env python
# encoding: utf-8
"""
docker_connect.py

Created by Mat Appelman on 2013-04-18.
Copyright (c) 2013 __MyCompanyName__. All rights reserved.
"""
import os
import sys
import zerorpc


c = zerorpc.Client()
c.connect("tcp://ec2-23-20-84-18.compute-1.amazonaws.com:4242")
print str(c)
for item in c.runCommand("sudo docker ps"):
    print item
