param(
    [string]$Manifest = "docs/evaluations/go-repo-pool-40.json",
    [string]$WorkRoot = "$env:TEMP/cervomut-go-pool-40",
    [string]$OutputRoot = "$env:TEMP/cervomut-tool-comparison-12",
    [string[]]$Names = @("cobra", "pflag", "moby", "hugo", "prometheus", "terraform", "grpc-go", "echo", "logrus", "validator", "decimal", "gjson"),
    [string[]]$Tools = @("cervomut", "gremlins", "gomu", "go-mutesting"),
    [int]$Workers = 2,
    [int]$GomuWorkers = 1,
    [int]$GoMutestingWorkers = 1,
    [int]$TimeoutSeconds = 600,
    [int]$MinFreeMemoryMB = 4096,
    [int]$MinFreeCommitMB = 8192,
    [int]$KillBelowFreeMemoryMB = 2048,
    [int]$KillBelowFreeCommitMB = 4096,
    [int]$MemoryWaitSeconds = 900,
    [int]$MemoryPollSeconds = 5,
    [switch]$Resume,
    [string]$CervoMutant = "$env:TEMP/cervomut-pool.exe",
    [string]$Gremlins = "$env:TEMP/cervomut-study-cobra/tools/gremlins.exe",
    [string]$Gomu = "$env:TEMP/cervomut-study-cobra/tools/gomu-patched.exe",
    [string]$GoMutesting = "$env:TEMP/cervomut-study-cobra/tools/go-mutesting-patched.exe"
)

$ErrorActionPreference = "Stop"

$manifestData = Get-Content -LiteralPath $Manifest -Raw | ConvertFrom-Json
$wanted = @{}
foreach ($name in $Names) {
    $wanted[$name] = $true
}
$repos = @($manifestData.repos | Where-Object { $wanted.ContainsKey($_.name) })
$wantedTools = @{}
foreach ($tool in $Tools) {
    $wantedTools[$tool] = $true
}

New-Item -ItemType Directory -Path $OutputRoot -Force | Out-Null

function Get-MemoryStatus {
    $physical = Get-CimInstance Win32_OperatingSystem
    $commit = Get-CimInstance Win32_PerfFormattedData_PerfOS_Memory
    $freeCommitMB = [int][math]::Round(($commit.CommitLimit - $commit.CommittedBytes) / 1MB, 0)
    return [pscustomobject]@{
        free_mb = [int][math]::Round($physical.FreePhysicalMemory / 1024, 0)
        free_commit_mb = $freeCommitMB
    }
}

function Wait-FreeMemory {
    param(
        [int]$MinFreeMemoryMB,
        [int]$MinFreeCommitMB,
        [int]$TimeoutSeconds
    )
    if ($MinFreeMemoryMB -le 0 -and $MinFreeCommitMB -le 0) {
        $memory = Get-MemoryStatus
        return [pscustomobject]@{ ready = $true; free_mb = $memory.free_mb; free_commit_mb = $memory.free_commit_mb; waited_seconds = 0 }
    }
    $sw = [Diagnostics.Stopwatch]::StartNew()
    while ($true) {
        $memory = Get-MemoryStatus
        $hasMemory = $MinFreeMemoryMB -le 0 -or $memory.free_mb -ge $MinFreeMemoryMB
        $hasCommit = $MinFreeCommitMB -le 0 -or $memory.free_commit_mb -ge $MinFreeCommitMB
        if ($hasMemory -and $hasCommit) {
            return [pscustomobject]@{ ready = $true; free_mb = $memory.free_mb; free_commit_mb = $memory.free_commit_mb; waited_seconds = [int]$sw.Elapsed.TotalSeconds }
        }
        if ($sw.Elapsed.TotalSeconds -ge $TimeoutSeconds) {
            return [pscustomobject]@{ ready = $false; free_mb = $memory.free_mb; free_commit_mb = $memory.free_commit_mb; waited_seconds = [int]$sw.Elapsed.TotalSeconds }
        }
        Start-Sleep -Seconds 15
    }
}

function Stop-ProcessTree {
    param([int]$ProcessId)
    $children = @(Get-CimInstance Win32_Process | Where-Object { $_.ParentProcessId -eq $ProcessId })
    foreach ($child in $children) {
        Stop-ProcessTree -ProcessId ([int]$child.ProcessId)
    }
    Stop-Process -Id $ProcessId -Force -ErrorAction SilentlyContinue
}

function Invoke-LoggedCommand {
    param(
        [string]$FilePath,
        [string[]]$Arguments,
        [string]$WorkingDirectory,
        [string]$LogPath,
        [int]$TimeoutSeconds,
        [int]$MinFreeMemoryMB,
        [int]$MinFreeCommitMB,
        [int]$KillBelowFreeMemoryMB,
        [int]$KillBelowFreeCommitMB,
        [int]$MemoryWaitSeconds,
        [int]$MemoryPollSeconds
    )
    $stdout = "$LogPath.stdout"
    $stderr = "$LogPath.stderr"
    Remove-Item -LiteralPath $stdout, $stderr, $LogPath -Force -ErrorAction SilentlyContinue
    $memory = Wait-FreeMemory -MinFreeMemoryMB $MinFreeMemoryMB -MinFreeCommitMB $MinFreeCommitMB -TimeoutSeconds $MemoryWaitSeconds
    if (!$memory.ready) {
        "skipped after waiting $($memory.waited_seconds)s for ${MinFreeMemoryMB}MB free memory and ${MinFreeCommitMB}MB free commit; free memory=$($memory.free_mb)MB free commit=$($memory.free_commit_mb)MB" | Set-Content -LiteralPath $LogPath
        return 125
    }
    try {
        $proc = Start-Process -FilePath $FilePath -ArgumentList $Arguments -WorkingDirectory $WorkingDirectory -NoNewWindow -PassThru -RedirectStandardOutput $stdout -RedirectStandardError $stderr
    } catch {
        $_ | Out-String | Set-Content -LiteralPath $LogPath
        return 125
    }
    $sw = [Diagnostics.Stopwatch]::StartNew()
    while (!$proc.HasExited) {
        if ($sw.Elapsed.TotalSeconds -ge $TimeoutSeconds) {
            Stop-ProcessTree -ProcessId $proc.Id
            [void]$proc.WaitForExit()
            "timed out after ${TimeoutSeconds}s" | Set-Content -LiteralPath $LogPath
            return 124
        }
        $memory = Get-MemoryStatus
        $belowMemory = $KillBelowFreeMemoryMB -gt 0 -and $memory.free_mb -lt $KillBelowFreeMemoryMB
        $belowCommit = $KillBelowFreeCommitMB -gt 0 -and $memory.free_commit_mb -lt $KillBelowFreeCommitMB
        if ($belowMemory -or $belowCommit) {
            Stop-ProcessTree -ProcessId $proc.Id
            [void]$proc.WaitForExit()
            "killed by memory watchdog after $([int]$sw.Elapsed.TotalSeconds)s; free memory=$($memory.free_mb)MB free commit=$($memory.free_commit_mb)MB" | Set-Content -LiteralPath $LogPath
            return 126
        }
        Start-Sleep -Seconds $MemoryPollSeconds
        $proc.Refresh()
    }
    if ($proc.ExitCode -eq 124) {
        "timed out after ${TimeoutSeconds}s" | Set-Content -LiteralPath $LogPath
        return 124
    }
    $parts = @()
    if (Test-Path -LiteralPath $stdout) { $parts += Get-Content -LiteralPath $stdout -Raw }
    if (Test-Path -LiteralPath $stderr) { $parts += Get-Content -LiteralPath $stderr -Raw }
    ($parts -join [Environment]::NewLine) | Set-Content -LiteralPath $LogPath
    Remove-Item -LiteralPath $stdout, $stderr -Force -ErrorAction SilentlyContinue
    $proc.Refresh()
    return [int]$proc.ExitCode
}

function Read-CervoReport($path) {
    if (!(Test-Path -LiteralPath $path)) { return @{} }
    $j = Get-Content -LiteralPath $path -Raw | ConvertFrom-Json
    return @{
        total = $j.summary.total
        killed = $j.summary.killed
        survived = $j.summary.survived
        not_covered = $j.summary.not_covered
        errors = $j.summary.compile_error
        timed_out = $j.summary.timed_out
        score = [math]::Round([double]$j.summary.score, 2)
    }
}

function Read-GremlinsReport($path) {
    if (!(Test-Path -LiteralPath $path)) { return @{} }
    $j = Get-Content -LiteralPath $path -Raw | ConvertFrom-Json
    return @{
        total = $j.mutants_total
        killed = $j.mutants_killed
        survived = $j.mutants_lived
        not_covered = $j.mutants_not_covered
        errors = 0
        timed_out = (($j.files | ForEach-Object { $_.mutations } | Where-Object { $_.status -eq "TIMED OUT" }).Count)
        score = [math]::Round([double]$j.test_efficacy, 2)
    }
}

function Read-GomuReport($path) {
    if (!(Test-Path -LiteralPath $path)) { return @{} }
    $j = Get-Content -LiteralPath $path -Raw | ConvertFrom-Json
    $groups = @{}
    foreach ($g in ($j.results | Group-Object status)) {
        $groups[$g.Name] = $g.Count
    }
    $killed = 0
    $survived = 0
    $errors = 0
    $notViable = 0
    if ($groups.ContainsKey("KILLED")) { $killed = [int]$groups["KILLED"] }
    if ($groups.ContainsKey("SURVIVED")) { $survived = [int]$groups["SURVIVED"] }
    if ($groups.ContainsKey("ERROR")) { $errors = [int]$groups["ERROR"] }
    if ($groups.ContainsKey("NOT_VIABLE")) { $notViable = [int]$groups["NOT_VIABLE"] }
    $denom = $killed + $survived
    $score = 0.0
    if ($denom -gt 0) { $score = [math]::Round(($killed / $denom) * 100, 2) }
    return @{
        total = $j.totalMutants
        killed = $killed
        survived = $survived
        not_covered = 0
        not_viable = $notViable
        errors = $errors
        timed_out = 0
        score = $score
    }
}

function Read-GoMutestingReport($path) {
    if (!(Test-Path -LiteralPath $path)) { return @{} }
    $j = Get-Content -LiteralPath $path -Raw | ConvertFrom-Json
    return @{
        total = $j.stats.totalMutantsCount
        killed = $j.stats.killedCount
        survived = $j.stats.escapedCount
        not_covered = $j.stats.notCoveredCount
        errors = $j.stats.errorCount
        timed_out = $j.stats.timeOutCount
        score = [math]::Round([double]$j.stats.msi * 100, 2)
    }
}

$results = @()
$summaryPath = Join-Path $OutputRoot "summary.json"
if ($Resume -and (Test-Path -LiteralPath $summaryPath)) {
    $loaded = @(Get-Content -LiteralPath $summaryPath -Raw | ConvertFrom-Json)
    foreach ($item in $loaded) {
        if ($item.PSObject.Properties.Name -contains "value") {
            $results += @($item.value)
        } elseif ($item.PSObject.Properties.Name -contains "repo") {
            $results += $item
        }
    }
}

function Has-Result($results, $repo, $tool) {
    return @($results | Where-Object { $_.repo -eq $repo -and $_.tool -eq $tool }).Count -gt 0
}

foreach ($repo in $repos) {
    $repoDir = Join-Path $WorkRoot $repo.name
    if (!(Test-Path -LiteralPath $repoDir)) {
        $results += [pscustomobject]@{ repo = $repo.name; tool = "all"; exit = 127; seconds = 0; note = "repo checkout missing" }
        continue
    }
    $repoOut = Join-Path $OutputRoot $repo.name
    New-Item -ItemType Directory -Path $repoOut -Force | Out-Null
    $toolDefs = New-Object System.Collections.Generic.List[object]
    $toolDefs.Add([pscustomobject]@{ name = "cervomut"; exe = $CervoMutant; args = @("run", $repo.target, "--profile", "gremlins-compatible", "--isolation", "overlay", "--workers", "$Workers", "--out", (Join-Path $repoOut "cervomut")); report = Join-Path $repoOut "cervomut/mutation-report.json"; parser = "cervo" }) | Out-Null
    $toolDefs.Add([pscustomobject]@{ name = "gremlins"; exe = $Gremlins; args = @("unleash", $repo.target, "--workers", "$Workers", "--threshold-efficacy", "0", "--threshold-mcover", "0", "--output", (Join-Path $repoOut "gremlins.json")); report = Join-Path $repoOut "gremlins.json"; parser = "gremlins" }) | Out-Null
    $toolDefs.Add([pscustomobject]@{ name = "gomu"; exe = $Gomu; args = @("run", $repo.target, "--workers", "$GomuWorkers", "--timeout", "$TimeoutSeconds", "--threshold", "0", "--fail-on-gate=false", "--output", "json"); report = Join-Path $repoDir "mutation-report.json"; parser = "gomu" }) | Out-Null
    $toolDefs.Add([pscustomobject]@{ name = "go-mutesting"; exe = $GoMutesting; args = @("/noop", "/quiet", "/no-diffs", "/logger-summary-json", "/logger-agentic-json", "/exec-timeout:$TimeoutSeconds", "/workers:$GoMutestingWorkers", $repo.target); report = Join-Path $repoDir "report.json"; parser = "go-mutesting" }) | Out-Null
    $selectedTools = @()
    foreach ($candidateTool in $toolDefs) {
        if ($wantedTools.ContainsKey($candidateTool.name)) {
            $selectedTools += $candidateTool
        }
    }
    foreach ($tool in $selectedTools) {
        if ($Resume -and (Has-Result $results $repo.name $tool.name)) {
            continue
        }
        Remove-Item -LiteralPath $tool.report -Force -ErrorAction SilentlyContinue
        $log = Join-Path $repoOut "$($tool.name).log"
        $sw = [Diagnostics.Stopwatch]::StartNew()
        $exit = Invoke-LoggedCommand -FilePath $tool.exe -Arguments $tool.args -WorkingDirectory $repoDir -LogPath $log -TimeoutSeconds $TimeoutSeconds -MinFreeMemoryMB $MinFreeMemoryMB -MinFreeCommitMB $MinFreeCommitMB -KillBelowFreeMemoryMB $KillBelowFreeMemoryMB -KillBelowFreeCommitMB $KillBelowFreeCommitMB -MemoryWaitSeconds $MemoryWaitSeconds -MemoryPollSeconds $MemoryPollSeconds
        $sw.Stop()
        $metrics = @{}
        switch ($tool.parser) {
            "cervo" { $metrics = Read-CervoReport $tool.report }
            "gremlins" { $metrics = Read-GremlinsReport $tool.report }
            "gomu" {
                if (Test-Path -LiteralPath $tool.report) {
                    Copy-Item -LiteralPath $tool.report -Destination (Join-Path $repoOut "gomu-mutation-report.json") -Force
                    $metrics = Read-GomuReport $tool.report
                }
            }
            "go-mutesting" {
                if (Test-Path -LiteralPath $tool.report) {
                    Copy-Item -LiteralPath $tool.report -Destination (Join-Path $repoOut "go-mutesting-report.json") -Force
                    $metrics = Read-GoMutestingReport $tool.report
                }
            }
        }
        $results += [pscustomobject]([ordered]@{
            repo = $repo.name
            target = $repo.target
            lane = $repo.lane
            domain = $repo.domain
            tool = $tool.name
            exit = $exit
            seconds = [math]::Round($sw.Elapsed.TotalSeconds, 2)
            total = $metrics.total
            killed = $metrics.killed
            survived = $metrics.survived
            not_covered = $metrics.not_covered
            not_viable = $metrics.not_viable
            errors = $metrics.errors
            timed_out = $metrics.timed_out
            score = $metrics.score
            log = $log
        })
        $results | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $summaryPath
    }
}

$results | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $summaryPath
$results | Format-Table -AutoSize
Write-Host "Summary: $summaryPath"
