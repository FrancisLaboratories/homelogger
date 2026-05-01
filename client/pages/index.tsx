'use client'

import React from 'react'
import { Container } from 'react-bootstrap'
import 'bootstrap-icons/font/bootstrap-icons.css'
import MyNavbar from '../components/Navbar'
import TasksDashboard from '../components/TasksDashboard'

const HomePage: React.FC = () => {
    return (
        <Container>
            <MyNavbar />
            <TasksDashboard />
        </Container>
    )
}

export default HomePage
