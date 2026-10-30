# Borderless Floating Timer Pill for Windows (PowerShell + WPF)
param(
    [string]$ServerUrl = "http://127.0.0.1:41890"
)

Add-Type -AssemblyName PresentationFramework, PresentationCore, WindowsBase

$xaml = @"
<Window xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation"
        xmlns:x="http://schemas.microsoft.com/winfx/2006/xaml"
        Title="Floating Timer"
        WindowStyle="None"
        AllowsTransparency="True"
        Topmost="True"
        ShowInTaskbar="False"
        Background="Transparent"
        SizeToContent="WidthAndHeight"
        ResizeMode="NoResize"
        FlowDirection="RightToLeft">
    <Border Name="Capsule"
            Background="#F00D121C"
            CornerRadius="22"
            BorderBrush="#33FFFFFF"
            BorderThickness="1"
            Padding="14,6,14,6"
            Margin="10">
        <Border.Effect>
            <DropShadowEffect Color="#000000" BlurRadius="20" ShadowDepth="4" Opacity="0.6"/>
        </Border.Effect>
        <StackPanel Orientation="Horizontal" VerticalAlignment="Center">
            <!-- Status Dot -->
            <Ellipse Name="StatusDot" Width="8" Height="8" Fill="#10B981" VerticalAlignment="Center" Margin="0,0,8,0"/>
            
            <!-- Project Label -->
            <TextBlock Name="ProjectText" Text="پروژه عمومی" FontSize="11" FontWeight="Bold" Foreground="#94A3B8" VerticalAlignment="Center" Margin="0,0,10,0" MaxWidth="90" TextTrimming="CharacterEllipsis"/>
            
            <!-- Clock -->
            <TextBlock Name="ClockText" Text="۰۰:۰۰:۰۰" FlowDirection="LeftToRight" FontSize="15" FontWeight="Bold" Foreground="#93C5FD" VerticalAlignment="Center" Margin="0,0,10,0"/>
            
            <!-- Pause / Resume Button -->
            <Button Name="PauseResumeBtn" Content="توقف ⏸" FontSize="11" FontWeight="Bold" Foreground="#FBBF24" Background="#33F59E0B" BorderBrush="#66F59E0B" Padding="8,3,8,3" Margin="0,0,6,0" Cursor="Hand">
                <Button.Resources>
                    <Style TargetType="Border">
                        <Setter Property="CornerRadius" Value="12"/>
                    </Style>
                </Button.Resources>
            </Button>
            
            <!-- Stop Button -->
            <Button Name="StopBtn" Content="پایان ⏹" FontSize="11" FontWeight="Bold" Foreground="#FB7185" Background="#33F43F5E" BorderBrush="#66F43F5E" Padding="8,3,8,3" Cursor="Hand">
                <Button.Resources>
                    <Style TargetType="Border">
                        <Setter Property="CornerRadius" Value="12"/>
                    </Style>
                </Button.Resources>
            </Button>
        </StackPanel>
    </Border>
</Window>
"@

$reader = [System.Xml.XmlReader]::Create([System.IO.StringReader]::new($xaml))
$window = [System.Windows.Markup.XamlReader]::Load($reader)

$capsule = $window.FindName("Capsule")
$dot = $window.FindName("StatusDot")
$projectText = $window.FindName("ProjectText")
$clockText = $window.FindName("ClockText")
$pauseResumeBtn = $window.FindName("PauseResumeBtn")
$stopBtn = $window.FindName("StopBtn")

# Make draggable
$capsule.Add_MouseLeftButtonDown({
    $window.DragMove()
})

$script:isPaused = $false
$script:isRunning = $false

function To-PersianDigits([string]$str) {
    $faDigits = @{'0'='۰';'1'='۱';'2'='۲';'3'='۳';'4'='۴';'5'='۵';'6'='۶';'7'='۷';'8'='۸';'9'='۹'}
    $res = ""
    foreach ($c in $str.ToCharArray()) {
        if ($faDigits.ContainsKey([string]$c)) { $res += $faDigits[[string]$c] } else { $res += $c }
    }
    return $res
}

function Format-Seconds([int]$sec) {
    $h = [math]::Floor($sec / 3600)
    $m = [math]::Floor(($sec % 3600) / 60)
    $s = $sec % 60
    return ("{0:D2}:{1:D2}:{2:D2}" -f $h, $m, $s)
}

function Update-State {
    try {
        $resp = Invoke-RestMethod -Uri "$ServerUrl/api/timer" -Method Get -TimeoutSec 2
        $timer = $resp.timer
        $elapsed = $resp.elapsed_seconds

        if (-not $timer.is_running) {
            $window.Close()
            return
        }

        $script:isRunning = $timer.is_running
        $script:isPaused = $timer.is_paused

        $projectText.Text = if ($timer.project) { $timer.project } else { "پروژه عمومی" }
        $clockText.Text = To-PersianDigits (Format-Seconds $elapsed)

        if ($script:isPaused) {
            $dot.Fill = [System.Windows.Media.Brushes]::Amber
            $pauseResumeBtn.Content = "ادامه ▶"
            $pauseResumeBtn.Foreground = [System.Windows.Media.BrushConverter]::new().ConvertFromString("#34D399")
            $pauseResumeBtn.Background = [System.Windows.Media.BrushConverter]::new().ConvertFromString("#3310B981")
        } else {
            $dot.Fill = [System.Windows.Media.BrushConverter]::new().ConvertFromString("#10B981")
            $pauseResumeBtn.Content = "توقف ⏸"
            $pauseResumeBtn.Foreground = [System.Windows.Media.BrushConverter]::new().ConvertFromString("#FBBF24")
            $pauseResumeBtn.Background = [System.Windows.Media.BrushConverter]::new().ConvertFromString("#33F59E0B")
        }
    } catch {
        # Keep running
    }
}

$pauseResumeBtn.Add_Click({
    $endpoint = if ($script:isPaused) { "/api/timer/resume" } else { "/api/timer/pause" }
    try {
        Invoke-RestMethod -Uri "$ServerUrl$endpoint" -Method Post -TimeoutSec 2 | Out-Null
        Update-State
    } catch {}
})

$stopBtn.Add_Click({
    try {
        Invoke-RestMethod -Uri "$ServerUrl/api/timer/stop" -Method Post -TimeoutSec 2 | Out-Null
    } catch {}
    $window.Close()
})

# Timer for polling
$dispatcherTimer = New-Object System.Windows.Threading.DispatcherTimer
$dispatcherTimer.Interval = [TimeSpan]::FromSeconds(1)
$dispatcherTimer.Add_Tick({ Update-State })
$dispatcherTimer.Start()

# Initial position: Top right
$screen = [System.Windows.SystemParameters]::WorkArea
$window.Left = $screen.Right - 320
$window.Top = $screen.Top + 60

Update-State
$window.ShowDialog() | Out-Null
