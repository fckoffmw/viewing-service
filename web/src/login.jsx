import {createRoot} from 'react-dom/client'
import AuthInput from './components/AuthInput'
import AuthSwitch from './components/AuthSwitch'
import AuthButton from './components/AuthButton'
import AuthLayout from './components/AuthLayout'

const LoginPage = () => {
    return (
        <AuthLayout title="Вход" id="login-form" footer={<AuthSwitch question="Ещё нет аккаунта?" href="/register.html" text="Регистрация"/>}>
            <AuthInput type="text" id="username" placeholder="Имя пользователя" minLength={3} autoComplete="username"/>
            <AuthInput type="password" id="password" placeholder="Пароль" minLength={4} autoComplete="current-password"/>
            <AuthButton text="Войти"/>
        </AuthLayout>
    )
}

createRoot(document.getElementById('root')).render(<LoginPage />)