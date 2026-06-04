$ErrorActionPreference = "Continue"
. (Join-Path $PSScriptRoot "generation-test-lib.ps1")

$outFile = Join-Path $PSScriptRoot "..\doc\development\gen-test-round2-results.json"

$ruleTests = @(
    @{ n=1; phrase="pwd"; wantIntent="current_dir" },
    @{ n=2; phrase="whoami"; wantIntent="current_user" },
    @{ n=3; phrase="date"; wantIntent="current_date" },
    @{ n=4; phrase="docker ps"; wantIntent="docker_ps" },
    @{ n=5; phrase="docker images"; wantIntent="docker_images" },
    @{ n=6; phrase="docker logs nginx"; wantIntent="docker_logs" },
    @{ n=7; phrase="grep error app.log"; wantIntent="search_text_in_file" },
    @{ n=8; phrase="search timeout in config.yaml"; wantIntent="search_text_in_file" },
    @{ n=9; phrase="git diff README.md"; wantIntent="git_diff_path" },
    @{ n=10; phrase="disk usage"; wantIntent="disk_usage" },
    @{ n=11; phrase="df ."; wantIntent="disk_usage" },
    @{ n=12; phrase="touch notes.txt"; wantIntent="create_empty_file" },
    @{ n=13; phrase="rmdir oldbackup"; wantIntent="remove_dir" },
    @{ n=14; phrase="current directory"; wantIntent="current_dir" },
    @{ n=15; phrase="which python"; wantIntent="which_command" },
    @{ n=16; phrase="echo Hello CLX"; wantIntent="print_text" },
    @{ n=17; phrase="env"; wantIntent="list_env" },
    @{ n=18; phrase="ss -tlnp"; wantIntent="netstat_listening" },
    @{ n=19; phrase="traceroute 1.1.1.1"; wantIntent="traceroute_host" },
    @{ n=20; phrase="find large files"; wantIntent="find_large_files" }
)

$nlTests = @(
    @{ n=1; phrase="list all running Windows services" },
    @{ n=2; phrase="show the computer hostname" },
    @{ n=3; phrase="flush the dns cache on this machine" },
    @{ n=4; phrase="display laptop battery status" },
    @{ n=5; phrase="show scheduled tasks that run today" },
    @{ n=6; phrase="list shared folders on this computer" },
    @{ n=7; phrase="show the name of the connected wifi network" },
    @{ n=8; phrase="extract a tar.gz file into the current folder" },
    @{ n=9; phrase="compare two folders and show differences" },
    @{ n=10; phrase="when did this pc last reboot" },
    @{ n=11; phrase="list usb devices currently connected" },
    @{ n=12; phrase="what is the default gateway" },
    @{ n=13; phrase="find which process is listening on port 443" },
    @{ n=14; phrase="show installed dotnet sdk versions" },
    @{ n=15; phrase="list globally installed npm packages" },
    @{ n=16; phrase="check if windows firewall is enabled" },
    @{ n=17; phrase="create a symbolic link to a directory" },
    @{ n=18; phrase="how much ram is installed on this machine" },
    @{ n=19; phrase="list files in the recycle bin" },
    @{ n=20; phrase="show only listening ports on localhost" }
)

Invoke-GenerationTestSuite -Round 2 -OutFile $outFile -RuleTests $ruleTests -NlTests $nlTests | Out-Null
