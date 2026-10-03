import AuthNav from "./AuthNav"

const AuthLayout = ({title, id, children, footer, onSubmit, error}) => {
    return (
         <>
            <AuthNav />
            <div className="auth-form">
                <h1>{title}</h1>
                <form id={id} onSubmit={onSubmit}>
                    {children}
                </form>
                {error && <div className="auth-error">{error}</div>}
                {footer}
            </div>
         </>
    )
}

export default AuthLayout