package command

import (
    "path/filepath"
    "net/http"
    "com.jfrog/bintray/cli/client"
    "os"
    "encoding/json"
    "io/ioutil"
    "log"
    "strconv"
)

type Upload struct {
    filePath string
}

type UploadArgs struct {
    Parallel uint32
    FilePath string
    Subject  string
    Repo     string
    Pkg      string
    Version  string
    Publish  bool
}

type UploadResult struct {
    FilePath string `json:"path"`
    Err      error  `json:"error,omitempty"`
    Message  string  `json:"message"`
}

func (res UploadResult) String() string {
    b, _ := json.Marshal(&res)
    return string(b)
}

type UploadHandle struct {
    fileCount int
    ch        chan *os.File
    results   []UploadResult
}

func (cmd Upload) Execute(bt *client.Bintray, args *UploadArgs) (result interface{}, err error) {
    //buffering - block sender until there is a listener
    ch := make(chan *os.File, args.Parallel)
    results := make([]UploadResult, 0)
    uh := &UploadHandle{0, ch, results}

    filePath := args.FilePath
    upload(filePath, bt, args, uh)

    for i := 0; i < uh.fileCount; i++ {
        <-uh.ch
    }

    /*log.Printf("RES: %s\n", uh.results)
    for _, res := range uh.results {
        log.Printf("RES: %s\n", res)
    }*/

    buf, _ := json.MarshalIndent(uh.results, "", "  ")
    log.Printf("%s\n", buf)

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
        //All writers block when full. Readers block when no value to read
        uh.fileCount++
        go func() {
            defer f.Close()
            log.Printf("Uploading: %v\n", filePath)
            res := uploadFile(f, bt, args)
            log.Printf("Uploaded: %s\n", res)
            uh.results = append(uh.results, *res)
            uh.ch <- f
        }()
    }
}

func uploadFile(f *os.File, bt *client.Bintray, args *UploadArgs) *UploadResult {
    //Use the relative path
    filePath := args.FilePath
    relPath, _ := filepath.Rel(filePath, f.Name())

    url := bt.ApiUrl + "content/" + args.Subject + "/" + args.Repo + "/" + args.Pkg +
    "/" + args.Version + "/" + relPath + "?publish=" + strconv.FormatBool(args.Publish)

    //log.Println("Uploading to: " + url)
    req, err := http.NewRequest("PUT", url, f)
    if err != nil {
        return &UploadResult{FilePath: f.Name(), Err: err}
    }
    updateRequestAuth(req, bt)

    client := &http.Client{}
    res, err := client.Do(req)
    if err != nil {
        return &UploadResult{FilePath: f.Name(), Err: err}
    }

    //fmt.Printf("RES: %v\n", res.Body)

    defer res.Body.Close()
    if res.StatusCode > 202 {
        log.Printf("Upload failed with: %v\n", res.Status)
    }

    body, err := ioutil.ReadAll(res.Body)
    perror(err)

    var vres map[string]string
    //fmt.Printf("BODY: %v\n", body)
    err = json.Unmarshal(body, &vres)

    //Artificial delay for tests - REMOVE
    //time.Sleep(time.Duration(rand.Intn(3000)))

    //log.Printf("MSG: %s\n", vres["message"])

    return &UploadResult{FilePath: f.Name(), Message: vres["message"], Err: err}
}
