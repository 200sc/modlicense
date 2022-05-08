# modlicense
A binary tool to track go.mod dependency versions

## Usage 

```bash
$ pwd
/my/go/project
$ modlicense 
{
    "module1 x.y.z": "MIT",
    "module2 x.y.z-alpha": "unknown"
}
$ # same as above
$ modlicense --directory . --file "go.mod" 
$ cat known-licenses.json
{
    "module2 x.y.z-alpha": "ECL"
}
$ modlicense --known=./known-licenses.json
{
    "module1 x.y.z": "MIT",
    "module2 x.y.z-alpha": "ECL"
}
$ # in automation
$ modlicense --known=./known-licenses.json > new-licenses.json
$ diff known-licenses.json new-licenses.json
$ # when there's no delta, proceed with CI
$ # when there is a delta, require manual verification to proceed
```
