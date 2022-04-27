# NavigaDoc Format

The purpose of this package is to define the NavigaDoc format as a Go struct and as a protobuf message, and provide functionality for commonly used operations for this format.

## Converter Functions

The navigadoc package has the following components

* Package github.com/navigacontentlab/navigadoc

      * contains utility functions for the NavigaDoc format
  
    
* Package github.com/navigacontentlab/navigadoc/doc
      
      * contains the golang definition of NavigaDoc

## Generate /doc and /rpc

* ./generate.sh

## TODO

* upgrade to go 1.18 (when linters have caught up)
* add more utility functions
* specify other formats related to NavigaDoc & Block
* go through example/test data and filter out what is not being used