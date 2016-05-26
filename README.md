# go-EFSM
EFSM Implementation for my thesis. This is by no means meant to be a generic EFSM implementation, use at your own risk.

# JSON Structure
From the root of the JSON there are three required keys:
  - info
  - states
  - functions

## Info
Info is a simple object with as required keys:
  - title (string)
  - version (string)

## States
States is an array of strings which contain state names.

## Functions
Functions is an array of function objects. These objects have the following required keys:
  - name (string)
  - transitions (array)
  
And the following optional key:
  - variable (string)

Where transitions is an array of transition objects, these transitions objects contain the following required keys:
  - from (string)
  - to (string)
these contain the names of the state to transition to, note: transitioning between the same state is allowed.

Variable is the name of the variable the can be set while calling the function.
