1. go get github.com/stellar/go

2. cd $GOPATH/src/github.com/stellar/go

ensure that you have Gopkg.lock and Gopkg.toml  
Although Gopkg.lock is supposed to be auto generated when one uses dep stellar has checked it in and Gopkg.lock and Gopkg.toml are not in sync.

3. Gopkg.lock has this entry given below which fails to resolve.  Removed it for the time being and it worked.  Since Gopkg.toml is out of sync didn't rely on it. 

[[projects]]
  branch = "default"
  digest = "1:24df057f15e7a09e75c1241cbe6f6590fd3eac9804f1110b02efade3214f042d"
  name = "bitbucket.org/ww/goautoneg"
  packages = ["."]
  pruneopts = "T"
  revision = "75cd24fc2f2c2a2088577d12123ddee5f54e0675"


4. dep ensure -v  // resolve all dependencies of stellar based on Gopkg.lock  .. 

This step populates the $PWD/vendor folder with all the dependencies 


5.  cp <marblesproject>/chaincode/src/marbles   to $GOPATH/src 

6.  cd $GOPATH/src/marbles 

7.  govendor add github.com/stellar/go/^    // to fetch stellar along with all dependencies.

    This populates the $GOPATH/src/marbles/vendor  folder with all the pkgs of stellar and its dependencies.

8.  import stellar pkgs in marbles chaincode and do a go build.
