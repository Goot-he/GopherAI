# GopherAI 图片识别模型下载脚本
# 下载 MobileNetV2 ONNX 模型、ImageNet 类别标签文件，以及 ONNX Runtime DLL

Write-Host "=== GopherAI 图片识别模型/依赖下载 ===" -ForegroundColor Cyan
Write-Host ""

$projectRoot = Split-Path -Parent $PSScriptRoot

# 创建 models 目录
$modelsDir = Join-Path $projectRoot "models"
if (-not (Test-Path $modelsDir)) {
    New-Item -ItemType Directory -Path $modelsDir -Force | Out-Null
    Write-Host "已创建目录: $modelsDir" -ForegroundColor Green
}

# 下载 MobileNetV2 ONNX 模型
$modelUrl = "https://github.com/onnx/models/raw/main/validated/vision/classification/mobilenet/model/mobilenetv2-7.onnx"
$modelPath = Join-Path $modelsDir "mobilenetv2-7.onnx"

if (-not (Test-Path $modelPath)) {
    Write-Host "正在下载 MobileNetV2 模型 (~14MB)..." -ForegroundColor Yellow
    try {
        Invoke-WebRequest -Uri $modelUrl -OutFile $modelPath -UseBasicParsing
        Write-Host "模型下载成功: $modelPath" -ForegroundColor Green
    }
    catch {
        Write-Host "从 GitHub 下载失败: $_" -ForegroundColor Red
        Write-Host "请手动下载模型文件并放置到 models/ 目录:" -ForegroundColor Yellow
        Write-Host "  https://github.com/onnx/models/raw/main/validated/vision/classification/mobilenet/model/mobilenetv2-7.onnx" -ForegroundColor Yellow
    }
}
else {
    Write-Host "模型文件已存在: $modelPath" -ForegroundColor Green
}

# 下载 ImageNet 类别标签
$labelsJsonUrl = "https://raw.githubusercontent.com/anishathalye/imagenet-simple-labels/master/imagenet-simple-labels.json"
$labelsJsonPath = Join-Path $modelsDir "imagenet-simple-labels.json"
$labelsTxtPath = Join-Path $modelsDir "imagenet_classes.txt"

if (-not (Test-Path $labelsTxtPath)) {
    Write-Host "正在下载 ImageNet 类别标签..." -ForegroundColor Yellow
    try {
        Invoke-WebRequest -Uri $labelsJsonUrl -OutFile $labelsJsonPath -UseBasicParsing
        Write-Host "标签 JSON 下载成功" -ForegroundColor Green
        
        $json = Get-Content $labelsJsonPath -Raw | ConvertFrom-Json
        $json | Out-File -FilePath $labelsTxtPath -Encoding UTF8
        Write-Host "标签文件已生成: $labelsTxtPath (共 $($json.Count) 个类别)" -ForegroundColor Green
    }
    catch {
        Write-Host "下载标签文件失败，请手动创建 imagenet_classes.txt" -ForegroundColor Red
    }
}
else {
    Write-Host "标签文件已存在: $labelsTxtPath" -ForegroundColor Green
}

# 下载 ONNX Runtime v1.22.0 DLL（与 go.mod 中 onnxruntime_go v1.22.0 匹配）
$ortVersion = "1.22.0"
$ortUrl = "https://github.com/microsoft/onnxruntime/releases/download/v$ortVersion/onnxruntime-win-x64-$ortVersion.zip"
$ortZipPath = Join-Path $projectRoot "onnxruntime-$ortVersion.zip"

Write-Host "---"
Write-Host "正在下载 ONNX Runtime v$ortVersion DLL..." -ForegroundColor Yellow
try {
    Invoke-WebRequest -Uri $ortUrl -OutFile $ortZipPath -UseBasicParsing
    Expand-Archive -Path $ortZipPath -DestinationPath $projectRoot -Force
    Copy-Item (Join-Path $projectRoot "onnxruntime-win-x64-$ortVersion\lib\*.dll") -Destination $projectRoot -Force
    Remove-Item $ortZipPath -ErrorAction SilentlyContinue
    Write-Host "ONNX Runtime v$ortVersion DLL 已安装到项目目录" -ForegroundColor Green
}
catch {
    Write-Host "下载 ONNX Runtime 失败: $_" -ForegroundColor Red
    Write-Host "请手动从以下地址下载并解压 onnxruntime.dll 和 onnxruntime_providers_shared.dll 到项目根目录:" -ForegroundColor Yellow
    Write-Host "  $ortUrl" -ForegroundColor Yellow
    Write-Host "注意: DLL 版本必须与 go.mod 中 onnxruntime_go 的版本一致 (当前: v$ortVersion)" -ForegroundColor Yellow
}

# 检查 DLL
Write-Host ""
Write-Host "=== 检查 ONNX Runtime DLL ===" -ForegroundColor Cyan
$dllPath = Join-Path $projectRoot "onnxruntime.dll"
if (Test-Path $dllPath) {
    $dllInfo = Get-Item $dllPath
    Write-Host "已找到: $dllPath ($($dllInfo.Length) bytes)" -ForegroundColor Green
}
else {
    Write-Host "未找到 onnxruntime.dll，请确保已下载并放置到项目根目录" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== 下载完成 ===" -ForegroundColor Cyan
