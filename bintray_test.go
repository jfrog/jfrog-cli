package main

import (
    "testing"
    "os"
    "io/ioutil"
    "fmt"
    "strconv"
    "log"
)

func TestSearchRepos(t *testing.T) {
    /*os.Clearenv()
    os.Setenv("BINTRAY_USER", "yoavl")
    os.Setenv("BINTRAY_KEY", "pass")*/
    os.Args = []string{"btray", "search-repos", "--subject", "jfrog"}
    main()
}

func TestUpload(t *testing.T) {
    tmpDir := os.TempDir() + "bintrayUploads/"
    os.Mkdir(tmpDir, 0777)

    //Deferred delete
    defer func() {
        f, err := os.Open(tmpDir)
        list, err := f.Readdir(-1)
        if err != nil {
            log.Fatalf("Cannot read dir: %v\n", err)
        }
        for _, file := range list {
            err = os.Remove(tmpDir + "/" + file.Name())
            if err != nil {
                fmt.Println(err)
            }
            //log.Println("Removed: " + file.Name())
        }
        err = os.Remove(tmpDir)
        if err != nil {
            fmt.Println(err)
        }
    }()

    for i := 1; i <= 15; i++ {
        file, _ := ioutil.TempFile(tmpDir, strconv.Itoa(i))
        fname := file.Name()
        fmt.Printf("Temp file: %s\n", fname)
    }
    os.Args = []string{"btray", "upload", "--path", tmpDir, "--repo", "gen", "--package", "pkg", "--version", "v2"}
    main()
}