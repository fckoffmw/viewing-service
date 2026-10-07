import { createRoot } from 'react-dom/client'
import AuthInput from './components/AuthInput'
import AuthSwitch from './components/AuthSwitch'
import AuthButton from './components/AuthButton'
import AuthLayout from './components/AuthLayout'
import { useState, useEffect } from 'react'

const RegisterPage = () => {
    const [username, setUsername] = useState('')
    const [newPassword, setNewPassword] = useState('')
    const [error, setError] = useState('')
    const [isLoading, setIsLoading] = useState(false)
    const buttonText = isLoading ? 'Регистрация...' : 'Зарегистрироваться'

    useEffect(() => {
        const checkAlreadyLogged = async () => {
            try {
                const res = await fetch('/auth/me')
                if (res.ok) {
                    window.location.href = '/index.html'
                }
            } catch {
                console.log('сервер не ответил')
            }
        }
        checkAlreadyLogged()
    }, [])

    const handleUsernameChange = (e) => {
        setUsername(e.target.value)
    }

    const handleNewPasswordChange = (e) => {
        setNewPassword(e.target.value)
    }

    const handleSubmit = async (e) => {
        e.preventDefault()
        setError('')
        setIsLoading(true)
        try {
            const res = await fetch('/auth/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username: username.trim(), password: newPassword }),
            })
            const data = await res.json()

            if (res.ok) {
                window.location.href = '/index.html'
                return
            }

            setError(data.error || 'Ошибка регистрации')
        } catch {
            setError('Ошибка соединения')
        } finally {
            setIsLoading(false)
        }
    }

    return (
        <AuthLayout onSubmit={handleSubmit} error={error} title="Регистрация" id="register-form" footer={<AuthSwitch question="Есть аккаунт?" href="/login.html" text="Вход" />}>
            <AuthInput value={username} onChange={handleUsernameChange} type="text" id="username" placeholder="Имя пользователя" minLength={3} autoComplete="username" />
            <AuthInput value={newPassword} onChange={handleNewPasswordChange} type="password" id="password" placeholder="Пароль" minLength={4} autoComplete="new-password" />
            <AuthButton text={buttonText} isLoading={isLoading} />
        </AuthLayout>
    )
}

createRoot(document.getElementById('root')).render(<RegisterPage />)