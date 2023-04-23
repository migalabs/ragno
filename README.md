# Ragno 

A simple yet-powerfull and updated crawler for Ethereum's EL.
It has some basics to crawl and connect to EL nodes.

# Installation
To install the tool:
```
make dependencies
make install
```

# Usage
At the moment, the tool only suports a small set of arguments:
```
USAGE:
   ragno [commands] [options...]

COMMANDS:
   discv4   discover4 prints nodes in the discovery4 network
   connect  connect and identify any given ENR
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

# Notes
Be careful with the input and output csv file names, as it will take `output.csv` by default on the `discv4` command. 

# Maintainer
@cortze

# Contributions
The code so far is a half implemented alpha - be patient. 
Feel free to suggest and to contribute though :)

