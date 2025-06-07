// 测试邮箱账号字段修复的脚本
// 这个脚本模拟浏览器扩展中的邮箱字段处理逻辑

console.log('🧪 开始测试邮箱账号字段修复...');

// 模拟当前账号数据
const mockCurrentAccount = {
    id: 1,
    platform_name: 'github.com',
    email_account_id: 5,
    email_address: 'dev@example.com',
    login_username: 'developer',
    phone_number: '+86 139****5678',
    notes: '开发者账号',
    created_at: '2024-01-01T00:00:00Z'
};

// 模拟编辑表单数据处理函数
function processEditFormData(formData, currentAccount) {
    console.log('📝 处理编辑表单数据:', formData);
    console.log('📋 当前账号数据:', currentAccount);
    
    const data = { ...formData };
    
    // 处理邮箱地址字段：将email_address转换为email_account_id
    if (data.email_address && data.email_address.trim() !== '') {
        // 如果邮箱地址与当前账号的邮箱地址相同，则使用当前的email_account_id
        if (data.email_address.trim() === (currentAccount.email_address || '')) {
            // 邮箱地址没有变化，使用当前的email_account_id
            if (currentAccount.email_account_id) {
                data.email_account_id = currentAccount.email_account_id;
            }
            console.log('✅ 邮箱地址未变化，使用现有email_account_id:', data.email_account_id);
        } else {
            // 邮箱地址发生了变化，需要查找或创建新的邮箱账户
            // 暂时不支持在编辑界面修改邮箱地址，显示错误信息
            return { 
                success: false, 
                error: '暂不支持在编辑界面修改邮箱地址，请删除后重新创建' 
            };
        }
    } else {
        // 邮箱地址为空，设置email_account_id为null
        data.email_account_id = null;
        console.log('✅ 邮箱地址为空，设置email_account_id为null');
    }
    
    // 移除email_address字段，因为后端更新API不接受这个字段
    delete data.email_address;
    
    console.log('📤 最终发送到后端的数据:', data);
    return { success: true, data };
}

// 测试用例
const testCases = [
    {
        name: '邮箱地址未变化',
        formData: {
            platform_name: 'github.com',
            email_address: 'dev@example.com',
            login_username: 'developer',
            phone_number: '+86 139****5678',
            notes: '开发者账号'
        },
        expected: {
            success: true,
            shouldHaveEmailAccountId: 5
        }
    },
    {
        name: '邮箱地址变化',
        formData: {
            platform_name: 'github.com',
            email_address: 'new@example.com',
            login_username: 'developer',
            phone_number: '+86 139****5678',
            notes: '开发者账号'
        },
        expected: {
            success: false,
            errorContains: '暂不支持在编辑界面修改邮箱地址'
        }
    },
    {
        name: '邮箱地址为空',
        formData: {
            platform_name: 'github.com',
            email_address: '',
            login_username: 'developer',
            phone_number: '+86 139****5678',
            notes: '开发者账号'
        },
        expected: {
            success: true,
            shouldHaveEmailAccountId: null
        }
    },
    {
        name: '邮箱地址只有空格',
        formData: {
            platform_name: 'github.com',
            email_address: '   ',
            login_username: 'developer',
            phone_number: '+86 139****5678',
            notes: '开发者账号'
        },
        expected: {
            success: true,
            shouldHaveEmailAccountId: null
        }
    },
    {
        name: '当前账号没有邮箱，表单也没有邮箱',
        formData: {
            platform_name: 'github.com',
            email_address: '',
            login_username: 'developer',
            phone_number: '+86 139****5678',
            notes: '开发者账号'
        },
        currentAccount: {
            ...mockCurrentAccount,
            email_account_id: null,
            email_address: ''
        },
        expected: {
            success: true,
            shouldHaveEmailAccountId: null
        }
    }
];

// 运行测试
let passedTests = 0;
let totalTests = testCases.length;

testCases.forEach((testCase, index) => {
    console.log(`\n🧪 测试 ${index + 1}: ${testCase.name}`);
    
    const currentAccount = testCase.currentAccount || mockCurrentAccount;
    const result = processEditFormData(testCase.formData, currentAccount);
    
    let passed = true;
    let errorMessage = '';
    
    // 检查成功/失败状态
    if (result.success !== testCase.expected.success) {
        passed = false;
        errorMessage = `期望 ${testCase.expected.success ? '成功' : '失败'}, 实际 ${result.success ? '成功' : '失败'}`;
    }
    
    // 如果期望成功，检查email_account_id
    if (testCase.expected.success && result.success) {
        if (testCase.expected.shouldHaveEmailAccountId !== undefined) {
            if (result.data.email_account_id !== testCase.expected.shouldHaveEmailAccountId) {
                passed = false;
                errorMessage = `期望email_account_id为 ${testCase.expected.shouldHaveEmailAccountId}, 实际为 ${result.data.email_account_id}`;
            }
        }
        
        // 检查是否正确移除了email_address字段
        if (result.data.hasOwnProperty('email_address')) {
            passed = false;
            errorMessage = '未正确移除email_address字段';
        }
    }
    
    // 如果期望失败，检查错误信息
    if (!testCase.expected.success && !result.success) {
        if (testCase.expected.errorContains && !result.error.includes(testCase.expected.errorContains)) {
            passed = false;
            errorMessage = `错误信息不包含期望的文本: "${testCase.expected.errorContains}"`;
        }
    }
    
    if (passed) {
        console.log(`✅ 通过`);
        if (result.success) {
            console.log(`   最终数据:`, result.data);
        } else {
            console.log(`   错误信息: ${result.error}`);
        }
        passedTests++;
    } else {
        console.log(`❌ 失败: ${errorMessage}`);
        console.log(`   结果:`, result);
    }
});

console.log(`\n📊 测试结果: ${passedTests}/${totalTests} 通过`);

if (passedTests === totalTests) {
    console.log('🎉 所有测试通过！邮箱账号字段修复成功！');
} else {
    console.log('⚠️ 部分测试失败，需要进一步检查。');
}

// 额外测试：验证后端期望的数据格式
console.log('\n🔍 后端期望的数据格式示例:');
const backendExpectedFormat = {
    login_username: 'developer',
    login_password: 'newpassword123', // 可选
    email_account_id: 5, // 或 null
    notes: '开发者账号',
    phone_number: '+86 139****5678'
    // 注意：没有 email_address 字段
};
console.log(JSON.stringify(backendExpectedFormat, null, 2));
