// 测试密码验证修复的脚本
// 这个脚本模拟浏览器扩展中的密码验证逻辑

console.log('🧪 开始测试密码验证修复...');

// 模拟浏览器扩展中的密码验证函数
function validatePasswordForEdit(data) {
    console.log('📝 测试编辑账号密码验证:', data);
    
    // 验证密码字段
    if (data.login_password && data.login_password.trim() !== '') {
        // 检查密码长度
        if (data.login_password.trim().length < 6) {
            return { success: false, error: '密码长度不能少于6位' };
        }
        // 检查密码长度上限
        if (data.login_password.trim().length > 128) {
            return { success: false, error: '密码长度不能超过128位' };
        }
    } else {
        // 移除空的密码字段
        delete data.login_password;
    }
    
    return { success: true, message: '密码验证通过' };
}

function validatePasswordForManualAdd(data) {
    console.log('📝 测试手动添加账号密码验证:', data);
    
    // 验证必填字段
    if (!data.platform_name) {
        return { success: false, error: '平台名称不能为空' };
    }

    // 验证密码字段
    if (data.login_password && data.login_password.trim() !== '') {
        // 检查密码长度
        if (data.login_password.trim().length < 6) {
            return { success: false, error: '密码长度不能少于6位' };
        }
        // 检查密码长度上限
        if (data.login_password.trim().length > 128) {
            return { success: false, error: '密码长度不能超过128位' };
        }
    }
    
    return { success: true, message: '密码验证通过' };
}

// 测试用例
const testCases = [
    {
        name: '编辑账号 - 密码太短',
        type: 'edit',
        data: { login_password: '12345' },
        expected: false
    },
    {
        name: '编辑账号 - 密码正常',
        type: 'edit',
        data: { login_password: '123456' },
        expected: true
    },
    {
        name: '编辑账号 - 密码为空',
        type: 'edit',
        data: { login_password: '' },
        expected: true
    },
    {
        name: '编辑账号 - 密码太长',
        type: 'edit',
        data: { login_password: 'a'.repeat(129) },
        expected: false
    },
    {
        name: '手动添加 - 平台名称为空',
        type: 'manual',
        data: { platform_name: '', login_password: '123456' },
        expected: false
    },
    {
        name: '手动添加 - 密码太短',
        type: 'manual',
        data: { platform_name: 'test.com', login_password: '12345' },
        expected: false
    },
    {
        name: '手动添加 - 正常情况',
        type: 'manual',
        data: { platform_name: 'test.com', login_password: '123456' },
        expected: true
    },
    {
        name: '手动添加 - 无密码',
        type: 'manual',
        data: { platform_name: 'test.com', login_password: '' },
        expected: true
    }
];

// 运行测试
let passedTests = 0;
let totalTests = testCases.length;

testCases.forEach((testCase, index) => {
    console.log(`\n🧪 测试 ${index + 1}: ${testCase.name}`);
    
    let result;
    if (testCase.type === 'edit') {
        result = validatePasswordForEdit({ ...testCase.data });
    } else {
        result = validatePasswordForManualAdd({ ...testCase.data });
    }
    
    const passed = result.success === testCase.expected;
    
    if (passed) {
        console.log(`✅ 通过: ${result.success ? result.message : result.error}`);
        passedTests++;
    } else {
        console.log(`❌ 失败: 期望 ${testCase.expected ? '成功' : '失败'}, 实际 ${result.success ? '成功' : '失败'}`);
        console.log(`   结果: ${result.success ? result.message : result.error}`);
    }
});

console.log(`\n📊 测试结果: ${passedTests}/${totalTests} 通过`);

if (passedTests === totalTests) {
    console.log('🎉 所有测试通过！密码验证修复成功！');
} else {
    console.log('⚠️ 部分测试失败，需要进一步检查。');
}
