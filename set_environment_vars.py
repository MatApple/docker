#!/usr/bin/env python
# encoding: utf-8
"""
set_environment_vars.py

Created by Mat Appelman on 2013-04-17.
Copyright (c) 2013 __MyCompanyName__. All rights reserved.
"""

import sys
import os


class set_environment_vars:
	def __init__(self):
		self.keys={
			"AWS_ACCESS_KEY_ID":"AKIAI7CYRNSO56TF3SBQ",
			"AWS_SECRET_ACCESS_KEY":"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",
			"AWS_KEYPAIR_NAME":"protobox-keypair",
			"AWS_SSH_PRIVKEY":"/Users/matappelman/aws/protobox-keypair.pem"
		}
	
	def set_keys(self):
		for key,val in self.keys.items():
			os.environ[key]=val
		return


if __name__ == '__main__':
	e=set_environment_vars()
	e.set_keys()