$ErrorActionPreference = "Continue"
. (Join-Path $PSScriptRoot "generation-test-lib.ps1")

$outFile = Join-Path $PSScriptRoot "..\doc\development\gen-test-round1-results.json"

$ruleTests = @(
    @{ n=1; phrase="find all files modified today"; wantIntent="find_modified_today" },
    @{ n=2; phrase="what is my ip"; wantIntent="show_ip_addresses" },
    @{ n=3; phrase="ipconfig"; wantIntent="show_ip_addresses" },
    @{ n=4; phrase="git status"; wantIntent="git_status" },
    @{ n=5; phrase="git log"; wantIntent="git_log" },
    @{ n=6; phrase="git diff"; wantIntent="git_diff" },
    @{ n=7; phrase="git branch"; wantIntent="git_branch_list" },
    @{ n=8; phrase="ping google.com"; wantIntent="ping_host" },
    @{ n=9; phrase="tracert google.com"; wantIntent="traceroute_host" },
    @{ n=10; phrase="ls ."; wantIntent="list_dir" },
    @{ n=11; phrase="cat README.md"; wantIntent="view_file" },
    @{ n=12; phrase="mkdir testdir"; wantIntent="make_dir" },
    @{ n=13; phrase="find file README.md"; wantIntent="find_file" },
    @{ n=14; phrase="head -n 10 README.md"; wantIntent="head_file" },
    @{ n=15; phrase="tail -n 5 README.md"; wantIntent="tail_file" },
    @{ n=16; phrase="curl -I https://example.com"; wantIntent="curl_url" },
    @{ n=17; phrase="netstat -an"; wantIntent="netstat_listening" },
    @{ n=18; phrase="del somefile.txt"; wantIntent="remove_file" },
    @{ n=19; phrase="copy a.txt b.txt"; wantIntent="copy_file" },
    @{ n=20; phrase="move a.txt b.txt"; wantIntent="move_file" }
)

$nlTests = @(
    @{ n=1; phrase="show me the 10 largest files in this folder" },
    @{ n=2; phrase="how much disk space is free" },
    @{ n=3; phrase="kill the process listening on port 8080" },
    @{ n=4; phrase="show running docker containers sorted by memory" },
    @{ n=5; phrase="compress this folder into a zip archive" },
    @{ n=6; phrase="show the last 50 lines of the newest log file" },
    @{ n=7; phrase="count how many lines of Go code are in this project" },
    @{ n=8; phrase="find every TODO comment in the source files" },
    @{ n=9; phrase="show my current branch and uncommitted changes" },
    @{ n=10; phrase="list all environment variables containing PATH" },
    @{ n=11; phrase="show CPU and memory usage right now" },
    @{ n=12; phrase="download a file from a URL and save it" },
    @{ n=13; phrase="find files larger than 100MB" },
    @{ n=14; phrase="show the git commit history for the last week" },
    @{ n=15; phrase="rename all .txt files to .md in this directory" },
    @{ n=16; phrase="show which program is using the most CPU" },
    @{ n=17; phrase="list installed python packages" },
    @{ n=18; phrase="show a file with line numbers" },
    @{ n=19; phrase="tail the windows event log" },
    @{ n=20; phrase="find duplicate files by name" }
)

Invoke-GenerationTestSuite -Round 1 -OutFile $outFile -RuleTests $ruleTests -NlTests $nlTests | Out-Null
