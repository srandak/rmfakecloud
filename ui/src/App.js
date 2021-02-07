// import 'bootstrap/dist/css/bootstrap.min.css'
// import React from 'react';
// import Nav from 'react-bootstrap/Nav'
// import Row from 'react-bootstrap/Row'
// import Navbar from 'react-bootstrap/Navbar'
// import Container from 'react-bootstrap/Container'
// import FileList from './components/FileList'

import React, { useEffect, useState } from 'react';
import { Router, Route, Link } from 'react-router-dom';

import { history, Role } from './_helpers';
import { authenticationService } from './_services/authentication.service';

import { PrivateRoute } from './_components/PrivateRoute';

import HomePage from './HomePage';
import AdminPage from './AdminPage';
import LoginPage from './LoginPage';

import { configureFakeBackend } from './_helpers';
configureFakeBackend();

export default function App() {

    const [authState, setAuthState] = useState({currentUser: authenticationService.currentUser, isAdmin: authenticationService.currentUser && authenticationService.currentUser.role === Role.Admin });

    useEffect(() => {

        authenticationService.currentUser.subscribe(x => setAuthState({currentUser: x, isAdmin: x && x.role === Role.Admin }))
    }, []);

    const logout = () => {
        authenticationService.logout();
        history.push('/login');
    }

    return (
        <Router history={ history }>
        <div>
            {authState.currentUser &&
                <nav className="navbar navbar-expand navbar-dark bg-dark">
                    <div className="navbar-nav">
                        <Link to="/" className="nav-item nav-link">Home</Link>
                        {authState.isAdmin && <Link to="/admin" className="nav-item nav-link">Admin</Link>}
                        <a onClick={logout} className="nav-item nav-link">Logout</a>
                    </div>
                </nav>
            }
            <div className="jumbotron">
                <div className="container">
                    <div className="row">
                        <div className="col-md-6 offset-md-3">
                            <PrivateRoute exact path="/" component={HomePage} />
                            <PrivateRoute path="/admin" roles={[Role.Admin]} component={AdminPage} />
                            <Route path="/login" component={LoginPage} />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </Router>
    );
}
