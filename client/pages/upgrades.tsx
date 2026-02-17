import React, { useEffect, useState } from 'react';
import { Button, Card, Col, Container, Form, Row, Table } from 'react-bootstrap';
import MyNavbar from '../components/Navbar';
import { SERVER_URL } from '@/pages/_app';

type BudgetCategory = {
  id: number;
  name: string;
};

type UpgradeProject = {
  id: number;
  title: string;
  description: string;
  status: string;
  priority: string;
  targetDate: string;
  estimatedCost: number;
  notes: string;
  categoryId?: number | null;
};

const UpgradesPage: React.FC = () => {
  const [projects, setProjects] = useState<UpgradeProject[]>([]);
  const [categories, setCategories] = useState<BudgetCategory[]>([]);
  const [newProject, setNewProject] = useState({
    title: '',
    description: '',
    status: 'planned',
    priority: 'medium',
    targetDate: '',
    estimatedCost: '',
    notes: '',
    categoryId: '',
  });

  const loadProjects = async () => {
    try {
      const resp = await fetch(`${SERVER_URL}/upgrades`);
      if (!resp.ok) return;
      const data = await resp.json();
      setProjects(data);
    } catch (e) {
      console.error('Error loading projects', e);
    }
  };

  const loadCategories = async () => {
    try {
      const resp = await fetch(`${SERVER_URL}/budget/categories`);
      if (!resp.ok) return;
      const data = await resp.json();
      setCategories(data);
    } catch (e) {
      console.error('Error loading categories', e);
    }
  };

  useEffect(() => {
    loadProjects();
    loadCategories();
  }, []);

  const handleAddProject = async () => {
    if (!newProject.title) return;
    const resp = await fetch(`${SERVER_URL}/upgrades/add`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        title: newProject.title,
        description: newProject.description,
        status: newProject.status,
        priority: newProject.priority,
        targetDate: newProject.targetDate,
        estimatedCost: Number(newProject.estimatedCost || 0),
        notes: newProject.notes,
        categoryId: newProject.categoryId ? Number(newProject.categoryId) : undefined,
      }),
    });
    if (!resp.ok) return;
    const created = await resp.json();
    setProjects((prev) => [...prev, created]);
    setNewProject({
      title: '',
      description: '',
      status: 'planned',
      priority: 'medium',
      targetDate: '',
      estimatedCost: '',
      notes: '',
      categoryId: '',
    });
  };

  const handleDeleteProject = async (id: number) => {
    const resp = await fetch(`${SERVER_URL}/upgrades/delete/${id}`, { method: 'DELETE' });
    if (!resp.ok) return;
    setProjects((prev) => prev.filter((p) => p.id !== id));
  };

  const categoryName = (id?: number | null) =>
    categories.find((c) => c.id === id)?.name || 'Uncategorized';

  return (
    <Container style={{ marginTop: '16px' }}>
      <MyNavbar />
      <h3>Upgrade Planning</h3>

      <Row className="g-3">
        <Col lg={12}>
          <Card>
            <Card.Body>
              <Table striped bordered hover size="sm">
                <thead>
                  <tr>
                    <th>Title</th>
                    <th>Status</th>
                    <th>Priority</th>
                    <th>Target Date</th>
                    <th>Estimated</th>
                    <th>Category</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  {projects.length === 0 ? (
                    <tr>
                      <td colSpan={7} style={{ textAlign: 'center' }}>No upgrade projects</td>
                    </tr>
                  ) : (
                    projects.map((p) => (
                      <tr key={p.id}>
                        <td>{p.title}</td>
                        <td>{p.status || '-'}</td>
                        <td>{p.priority || '-'}</td>
                        <td>{p.targetDate || '-'}</td>
                        <td>${Number(p.estimatedCost || 0).toFixed(2)}</td>
                        <td>{categoryName(p.categoryId)}</td>
                        <td style={{ width: 70 }}>
                          <Button variant="outline-danger" size="sm" onClick={() => handleDeleteProject(p.id)}>Delete</Button>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </Table>

              <Form className="mt-3">
                <Row className="g-2">
                  <Col md={4}>
                    <Form.Control placeholder="Project title" value={newProject.title} onChange={(e) => setNewProject({ ...newProject, title: e.target.value })} />
                  </Col>
                  <Col md={3}>
                    <Form.Select value={newProject.status} onChange={(e) => setNewProject({ ...newProject, status: e.target.value })}>
                      <option value="planned">Planned</option>
                      <option value="in-progress">In Progress</option>
                      <option value="completed">Completed</option>
                      <option value="on-hold">On Hold</option>
                    </Form.Select>
                  </Col>
                  <Col md={3}>
                    <Form.Select value={newProject.priority} onChange={(e) => setNewProject({ ...newProject, priority: e.target.value })}>
                      <option value="low">Low</option>
                      <option value="medium">Medium</option>
                      <option value="high">High</option>
                    </Form.Select>
                  </Col>
                  <Col md={2}>
                    <Form.Control type="date" value={newProject.targetDate} onChange={(e) => setNewProject({ ...newProject, targetDate: e.target.value })} />
                  </Col>
                  <Col md={2}>
                    <Form.Control type="number" step="0.01" placeholder="Estimate" value={newProject.estimatedCost} onChange={(e) => setNewProject({ ...newProject, estimatedCost: e.target.value })} />
                  </Col>
                  <Col md={4}>
                    <Form.Select value={newProject.categoryId} onChange={(e) => setNewProject({ ...newProject, categoryId: e.target.value })}>
                      <option value="">Category</option>
                      {categories.map((c) => (
                        <option key={c.id} value={c.id}>{c.name}</option>
                      ))}
                    </Form.Select>
                  </Col>
                  <Col md={6}>
                    <Form.Control placeholder="Description" value={newProject.description} onChange={(e) => setNewProject({ ...newProject, description: e.target.value })} />
                  </Col>
                  <Col md={6}>
                    <Form.Control placeholder="Notes" value={newProject.notes} onChange={(e) => setNewProject({ ...newProject, notes: e.target.value })} />
                  </Col>
                  <Col md={12}>
                    <Button onClick={handleAddProject}>Add Project</Button>
                  </Col>
                </Row>
              </Form>
            </Card.Body>
          </Card>
        </Col>
      </Row>
    </Container>
  );
};

export default UpgradesPage;
