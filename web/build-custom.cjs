#!/usr/bin/env node

/**
 * 自定义构建脚本
 * 支持指定输出目录
 * 使用方法：node build-custom.cjs [OUTPUT_DIR]
 * 示例：node build-custom.cjs ../deploy/frontend
 */

const { spawn } = require('child_process');
const path = require('path');

// 获取命令行参数
const outputDir = process.argv[2] || 'dist';
const emptyOutDir = process.argv.includes('--empty') || process.argv.includes('--emptyOutDir');

console.log(`🚀 开始构建项目...`);
console.log(`📁 输出目录: ${outputDir}`);
if (emptyOutDir) {
  console.log(`🗑️  将清空输出目录`);
}

// 设置环境变量
const env = {
  ...process.env,
  VITE_OUT_DIR: outputDir
};

// 构建 Vite 命令参数
const viteArgs = ['run', 'build'];
if (emptyOutDir) {
  // 需要直接调用 vite 命令来传递 --emptyOutDir 参数
  const buildProcess = spawn('npx', ['vite', 'build', '--mode', 'production', '--emptyOutDir'], {
    env,
    stdio: 'inherit',
    shell: true
  });
  
  buildProcess.on('close', (code) => {
    if (code === 0) {
      console.log(`✅ 构建完成！`);
      console.log(`📍 文件已输出到: ${path.resolve(outputDir)}`);
    } else {
      console.error(`❌ 构建失败，退出码: ${code}`);
      process.exit(code);
    }
  });

  buildProcess.on('error', (error) => {
    console.error(`❌ 构建过程出错: ${error.message}`);
    process.exit(1);
  });
  
  return;
}

// 执行构建命令
const buildProcess = spawn('npm', viteArgs, {
  env,
  stdio: 'inherit',
  shell: true
});

buildProcess.on('close', (code) => {
  if (code === 0) {
    console.log(`✅ 构建完成！`);
    console.log(`📍 文件已输出到: ${path.resolve(outputDir)}`);
  } else {
    console.error(`❌ 构建失败，退出码: ${code}`);
    process.exit(code);
  }
});

buildProcess.on('error', (error) => {
  console.error(`❌ 构建过程出错: ${error.message}`);
  process.exit(1);
});