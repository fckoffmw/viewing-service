import {createRoot} from 'react-dom/client'
import AuthInput from './components/AuthInput'
import AuthSwitch from './components/AuthSwitch'
import AuthButton from './components/AuthButton'
import AuthLayout from './components/AuthLayout'

const RegisterPage = () => {
    return (
        <AuthLayout title="Регистрация" id="register-form" footer={<AuthSwitch question="Есть аккаунт?" href="/login.html" text="Вход"/>}>
            <AuthInput type="text" id="username" placeholder="Имя пользователя" minLength={3} autoComplete="username"/>
            <AuthInput type="password" id="password" placeholder="Пароль" minLength={4} autoComplete="new-password"/>
            <AuthButton text="Зарегистрироваться"/>
        </AuthLayout>
    )
}

createRoot(document.getElementById('root')).render(<RegisterPage />)