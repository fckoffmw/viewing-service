const AuthButton = ({ text, isLoading }) => {
    return (
        <button type="submit" className="auth-btn" disabled={isLoading}>{text}</button>
    )
}

export default AuthButton