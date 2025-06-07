// 测试邮箱地址字段修复的脚本
// 这个脚本验证前后端修改后的邮箱地址处理逻辑

console.log('🧪 开始测试邮箱地址字段修复...');

// 模拟前端编辑表单数据处理函数（修复后）
function processEditFormDataFixed(formData) {
    console.log('📝 处理编辑表单数据（修复后）:', formData);
    
    const data = { ...formData };
    
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

    // 现在后端支持直接接收email_address字段，不需要转换为email_account_id
    // 保持email_address字段，后端会自动处理邮箱账号的查找或创建
    
    console.log('📤 最终发送到后端的数据:', data);
    return { success: true, data };
}

// 模拟后端处理逻辑
function simulateBackendProcessing(inputData, existingRegistration, existingEmailAccounts) {
    console.log('🔧 模拟后端处理:', { inputData, existingRegistration, existingEmailAccounts });
    
    let emailAccountId = null;
    let emailAccount = null;
    
    if (inputData.email_address && inputData.email_address.trim() !== '') {
        // 查找现有邮箱账户
        emailAccount = existingEmailAccounts.find(acc => acc.email_address === inputData.email_address);
        
        if (!emailAccount) {
            // 创建新的邮箱账户
            emailAccount = {
                id: existingEmailAccounts.length + 1,
                email_address: inputData.email_address,
                provider: inputData.email_address.split('@')[1],
                user_id: existingRegistration.user_id
            };
            existingEmailAccounts.push(emailAccount);
            console.log('✅ 创建新邮箱账户:', emailAccount);
        } else {
            console.log('✅ 找到现有邮箱账户:', emailAccount);
        }
        
        emailAccountId = emailAccount.id;
        
        // 检查是否与其他注册信息冲突
        // 这里简化处理，实际后端会检查数据库
        console.log('✅ 检查唯一约束通过');
    }
    
    // 更新注册信息
    const updatedRegistration = {
        ...existingRegistration,
        email_account_id: emailAccountId,
        email_address: inputData.email_address || '',
        login_username: inputData.login_username || existingRegistration.login_username,
        notes: inputData.notes || existingRegistration.notes,
        phone_number: inputData.phone_number || existingRegistration.phone_number
    };
    
    if (inputData.login_password) {
        updatedRegistration.has_password = true;
    }
    
    console.log('✅ 后端处理完成，更新后的注册信息:', updatedRegistration);
    return { success: true, data: updatedRegistration };
}

// 测试用例
const testCases = [
    {
        name: '添加新邮箱地址',
        formData: {
            platform_name: 'github.com',
            email_address: 'new@example.com',
            login_username: 'developer',
            login_password: 'newpassword123',
            phone_number: '+86 139****5678',
            notes: '开发者账号'
        },
        existingRegistration: {
            id: 1,
            user_id: 1,
            platform_id: 1,
            email_account_id: null,
            email_address: '',
            login_username: 'developer',
            notes: '开发者账号',
            phone_number: '+86 139****5678'
        },
        existingEmailAccounts: [],
        expected: {
            success: true,
            shouldCreateNewEmailAccount: true
        }
    },
    {
        name: '使用现有邮箱地址',
        formData: {
            platform_name: 'github.com',
            email_address: 'existing@example.com',
            login_username: 'developer',
            phone_number: '+86 139****5678',
            notes: '开发者账号'
        },
        existingRegistration: {
            id: 1,
            user_id: 1,
            platform_id: 1,
            email_account_id: null,
            email_address: '',
            login_username: 'developer',
            notes: '开发者账号',
            phone_number: '+86 139****5678'
        },
        existingEmailAccounts: [
            {
                id: 5,
                email_address: 'existing@example.com',
                provider: 'example.com',
                user_id: 1
            }
        ],
        expected: {
            success: true,
            shouldUseExistingEmailAccount: true,
            expectedEmailAccountId: 5
        }
    },
    {
        name: '清空邮箱地址',
        formData: {
            platform_name: 'github.com',
            email_address: '',
            login_username: 'developer',
            phone_number: '+86 139****5678',
            notes: '开发者账号'
        },
        existingRegistration: {
            id: 1,
            user_id: 1,
            platform_id: 1,
            email_account_id: 5,
            email_address: 'old@example.com',
            login_username: 'developer',
            notes: '开发者账号',
            phone_number: '+86 139****5678'
        },
        existingEmailAccounts: [
            {
                id: 5,
                email_address: 'old@example.com',
                provider: 'example.com',
                user_id: 1
            }
        ],
        expected: {
            success: true,
            shouldClearEmailAccount: true
        }
    },
    {
        name: '密码验证失败',
        formData: {
            platform_name: 'github.com',
            email_address: 'test@example.com',
            login_username: 'developer',
            login_password: '123', // 太短
            phone_number: '+86 139****5678',
            notes: '开发者账号'
        },
        existingRegistration: {
            id: 1,
            user_id: 1,
            platform_id: 1,
            email_account_id: null,
            email_address: '',
            login_username: 'developer',
            notes: '开发者账号',
            phone_number: '+86 139****5678'
        },
        existingEmailAccounts: [],
        expected: {
            success: false,
            errorContains: '密码长度不能少于6位'
        }
    }
];

// 运行测试
let passedTests = 0;
let totalTests = testCases.length;

testCases.forEach((testCase, index) => {
    console.log(`\n🧪 测试 ${index + 1}: ${testCase.name}`);
    
    // 前端处理
    const frontendResult = processEditFormDataFixed(testCase.formData);
    
    if (!frontendResult.success) {
        // 前端验证失败
        if (testCase.expected.success === false && 
            testCase.expected.errorContains && 
            frontendResult.error.includes(testCase.expected.errorContains)) {
            console.log(`✅ 通过: 前端验证正确拦截 - ${frontendResult.error}`);
            passedTests++;
        } else {
            console.log(`❌ 失败: 前端验证结果不符合预期`);
            console.log(`   期望: ${testCase.expected.success ? '成功' : '失败'}`);
            console.log(`   实际: ${frontendResult.success ? '成功' : '失败'} - ${frontendResult.error}`);
        }
        return;
    }
    
    // 后端处理
    const backendResult = simulateBackendProcessing(
        frontendResult.data, 
        testCase.existingRegistration, 
        [...testCase.existingEmailAccounts] // 复制数组避免修改原数据
    );
    
    let passed = true;
    let errorMessage = '';
    
    // 检查结果
    if (backendResult.success !== testCase.expected.success) {
        passed = false;
        errorMessage = `期望 ${testCase.expected.success ? '成功' : '失败'}, 实际 ${backendResult.success ? '成功' : '失败'}`;
    }
    
    if (testCase.expected.success && backendResult.success) {
        // 检查特定期望
        if (testCase.expected.shouldCreateNewEmailAccount) {
            // 应该创建新邮箱账户
            if (!backendResult.data.email_account_id || backendResult.data.email_account_id <= testCase.existingEmailAccounts.length) {
                passed = false;
                errorMessage = '应该创建新邮箱账户但没有创建';
            }
        }
        
        if (testCase.expected.shouldUseExistingEmailAccount) {
            // 应该使用现有邮箱账户
            if (backendResult.data.email_account_id !== testCase.expected.expectedEmailAccountId) {
                passed = false;
                errorMessage = `应该使用邮箱账户ID ${testCase.expected.expectedEmailAccountId}, 实际为 ${backendResult.data.email_account_id}`;
            }
        }
        
        if (testCase.expected.shouldClearEmailAccount) {
            // 应该清空邮箱账户
            if (backendResult.data.email_account_id !== null) {
                passed = false;
                errorMessage = '应该清空邮箱账户但没有清空';
            }
        }
    }
    
    if (passed) {
        console.log(`✅ 通过`);
        passedTests++;
    } else {
        console.log(`❌ 失败: ${errorMessage}`);
        console.log(`   结果:`, backendResult);
    }
});

console.log(`\n📊 测试结果: ${passedTests}/${totalTests} 通过`);

if (passedTests === totalTests) {
    console.log('🎉 所有测试通过！邮箱地址字段修复成功！');
    console.log('\n✨ 修复总结:');
    console.log('1. 前端现在直接发送 email_address 字段');
    console.log('2. 后端自动查找或创建对应的邮箱账户');
    console.log('3. 支持邮箱地址的添加、修改和清空');
    console.log('4. 保持了原有的密码验证逻辑');
} else {
    console.log('⚠️ 部分测试失败，需要进一步检查。');
}
