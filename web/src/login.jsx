import { createRoot } from 'react-dom/client'
import AuthInput from './components/AuthInput'
import AuthSwitch from './components/AuthSwitch'
import AuthButton from './components/AuthButton'
import AuthLayout from './components/AuthLayout'
import { useState } from 'react'

const LoginPage = () => {
    const [username, setUsername] = useState('')
    const [password, setPassword] = useState('')
    const [error, setError] = useState('')
    const [isLoading, setIsLoading] = useState(false)
    const isButtonLoad = isLoading ? 'Вход...' : 'Войти'

    const handleUsernameChange = (e) => {
        setUsername(e.target.value)
    }

    const handlePasswordChange = (e) => {
        setPassword(e.target.value)
    }

    const handleSubmit = async (e) => {
        e.preventDefault()
        setError('')
        setIsLoading(true)
        try {
            const res = await fetch('/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username: username.trim(), password: password }),
            })
            const data = await res.json()

            if (res.ok) {
                window.location.href = '/index.html'
                return
            }

            setError(data.error || 'Ошибка входа')
        } catch {
            setError('Ошибка соединения')
        } finally {
            setIsLoading(false)
        }
    }

    return (
        <AuthLayout onSubmit={handleSubmit} error={error} title="Вход" id="login-form" footer={<AuthSwitch question="Ещё нет аккаунта?" href="/register.html" text="Регистрация" />}>
            <AuthInput value={username} onChange={handleUsernameChange} type="text" id="username" placeholder="Имя пользователя" minLength={3} autoComplete="username" />
            <AuthInput value={password} onChange={handlePasswordChange} type="password" id="password" placeholder="Пароль" minLength={4} autoComplete="current-password" />
            <AuthButton text={isButtonLoad} isLoading={isLoading} />
        </AuthLayout>
    )
}

createRoot(document.getElementById('root')).render(<LoginPage />)