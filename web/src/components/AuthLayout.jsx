import AuthNav from "./AuthNav"

const AuthLayout = ({title, id, children, footer}) => {
    return (
         <>
            <AuthNav />
            <div className="auth-form">
                <h1>{title}</h1>
                <form id={id}>
                    {children}
                </form>
                <div id="error" className="auth-error"></div>
                {footer}
            </div>
         </>
    )
}

export default AuthLayout