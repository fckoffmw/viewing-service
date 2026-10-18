import { createRoot } from 'react-dom/client'
import AuthInput from './components/AuthInput'
import AuthSwitch from './components/AuthSwitch'
import AuthButton from './components/AuthButton'
import AuthLayout from './components/AuthLayout'
import useRedirectIfLoggedIn from './hooks/useRedirectIfLoggedIn'
import useAuthForm from './hooks/useAuthForm'

const LoginPage = () => {
    useRedirectIfLoggedIn()

    const {
        username,
        password,
        error,
        isLoading,
        handlePasswordChange,
        handleUsernameChange,
        handleSubmit,
    } = useAuthForm('/auth/login', 'Ошибка входа')
    const buttonText = isLoading ? 'Вход...' : 'Войти'

    return (
        <AuthLayout onSubmit={handleSubmit} error={error} title="Вход" id="login-form" footer={<AuthSwitch question="Ещё нет аккаунта?" href="/register.html" text="Регистрация" />}>
            <AuthInput value={username} onChange={handleUsernameChange} type="text" id="username" placeholder="Имя пользователя" minLength={3} autoComplete="username" />
            <AuthInput value={password} onChange={handlePasswordChange} type="password" id="password" placeholder="Пароль" minLength={4} autoComplete="current-password" />
            <AuthButton text={buttonText} isLoading={isLoading} />
        </AuthLayout>
    )
}

createRoot(document.getElementById('root')).render(<LoginPage />)