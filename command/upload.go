package command

import (
    "path/filepath"
    "net/http"
    "com.jfrog/bintray/cli/client"
    "os"
    "encoding/json"
    "io/ioutil"
    "log"
    "fmt"
    "time"
    "strconv"
)

type Upload struct {
    filePath string
}

type UploadArgs struct {
    Parallel uint32
    FilePath string
    Subject string
    Repo string
    Pkg string
    Version string
    Publish bool
}

type UploadResult struct {
    filePath string
    err      error
    json     string
}

func (res UploadResult) String() string {
    return fmt.Sprintf("path: %s, err: %v, json: %s", res.filePath, res.err, res.json)
}

//TODO: create a CommandArgs using NewArgs() and use it in execute()

type UploadHandle struct {
    ch      chan *os.File
    results []*UploadResult
}

func (cmd Upload) Execute(bt *client.Bintray, args *UploadArgs) (result interface{}, err error) {
    //buffering - block sender until there is a listener
    ch := make(chan *os.File, args.Parallel)
    results := make([]*UploadResult, cap(ch))
    uh := &UploadHandle{ch, results}

    filePath := args.FilePath
    upload(filePath, bt, args, uh)

    buf, _ := json.MarshalIndent(uh.results, "", "  ")
    fmt.Printf("%s\n", buf)

    //Todo: collect all results to an array and return it
    return uh.results, nil
}

func upload(filePath string, bt *client.Bintray, args *UploadArgs, uh *UploadHandle) {
    f, err := os.Open(filePath)
    if err != nil {
        log.Fatalf("Cannot open file: %s\n", filePath)
    }

    fi, _ := f.Stat()
    if fi.IsDir() {
        //fmt.Println("*** DIR: " + f.Name())

        //Recurse through children
        list, err := f.Readdirnames(-1)
        if err != nil {
            log.Fatalf("Cannot read dir names: %v\n", err)
        }
        //        fmt.Println("len:", len(list))

        for _, child := range list {
            //fmt.Println("*** CHILD: " + child)
            upload(filePath + child, bt, args, uh)
        }
    } else {
        //        fmt.Println("*** FILE: " + fi.Name())
        uh.ch <- f
        go func() {
            defer f.Close()
            log.Printf("Uploading: %v\n", filePath)
            res := uploadFile(f, bt, args)
            uh.results = append(uh.results, res)
            fmt.Printf("Upload done for %s (count: %d)\n", res.filePath, len(uh.results))
            <-uh.ch
        }()
    }
}

func uploadFile(f *os.File, bt *client.Bintray, args *UploadArgs) *UploadResult {
    //Use the relative path
    filePath := args.FilePath
    relPath, _ := filepath.Rel(filePath, f.Name())

    url := bt.ApiUrl + "content/" + args.Subject + "/" + args.Repo + "/" + args.Pkg +
    "/" + args.Version + "/" + relPath + "?publish=" + strconv.FormatBool(args.Publish)

//    log.Println("Uploading to: " + url)
    req, err := http.NewRequest("PUT", url, f)
    if err != nil {
        return &UploadResult{filePath: f.Name(), err: err}
    }
    updateRequestAuth(req, bt)

    client := &http.Client{}
    res, err := client.Do(req)
    if err != nil {
        return &UploadResult{filePath: f.Name(), err: err}
    }

    //    fmt.Printf("RES: %v", res)

    defer res.Body.Close()
    if res.StatusCode > 202 {
        log.Printf("Upload failed with: %v\n", res.Status)
    }

    body, err := ioutil.ReadAll(res.Body)
    perror(err)
    var vres map[string]string
    //    fmt.Printf("BODY: %s", json: vres["message"])
    err = json.Unmarshal(body, &vres)

    //REMOVE!!!
    time.Sleep(time.Millisecond * 2000)//time.Duration(rand.Intn(1000))

    return &UploadResult{filePath: f.Name(), json: vres["message"], err: err}
}
