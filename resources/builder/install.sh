#!/bin/bash

###########################################################
#
# Copyright (c) 2018 codeliveroil. All rights reserved.
#
# This work is licensed under the terms of the MIT license.
# For a copy, see <https://opensource.org/licenses/MIT>.
#
###########################################################

name="pping"
help_opt=" -h"
path="/usr/local/bin"

mkdir -p /usr/local/bin
if [ $(echo "$PATH" | grep "${path}") == "" ]; then
  echo "### Added by codeliveroil ###" >> ~/.bash_profile
  echo "export PATH=\$PATH:${path}" >> ~/.bash_profile
fi
cp ${name} ${path}

if [ $? -ne 0 ]; then
  echo "Installation was unsuccessful. Maybe you don't have permissions to write to ${path}. Try copying '${name}' to ${path} manually."
  exit 1
fi

echo "Installation successful. Run '${name}${help_opt}' for usage."
