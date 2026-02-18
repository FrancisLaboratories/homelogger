import React, { useEffect, useState } from 'react';
import { Button, Card, Col, Container, Form, Row, Table, Toast, ToastContainer } from 'react-bootstrap';
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
  const [editProjectId, setEditProjectId] = useState<number | null>(null);
  const [editProject, setEditProject] = useState<UpgradeProject | null>(null);
  const [projectError, setProjectError] = useState<string>('');
  const [notice, setNotice] = useState<{ variant: 'success' | 'danger'; message: string } | null>(null);
  const showNotice = (variant: 'success' | 'danger', message: string) => setNotice({ variant, message });

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
    if (!newProject.title) {
      setProjectError('Project title is required.');
      return;
    }
    setProjectError('');
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
    if (!resp.ok) {
      showNotice('danger', 'Failed to add project.');
      return;
    }
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
    showNotice('success', 'Project added.');
  };

  const handleStartEdit = (project: UpgradeProject) => {
    setEditProjectId(project.id);
    setEditProject({ ...project });
  };

  const handleCancelEdit = () => {
    setEditProjectId(null);
    setEditProject(null);
  };

  const handleSaveEdit = async () => {
    if (!editProject || !editProjectId) return;
    if (!editProject.title) {
      setProjectError('Project title is required.');
      return;
    }
    const resp = await fetch(`${SERVER_URL}/upgrades/update/${editProjectId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        title: editProject.title,
        description: editProject.description,
        status: editProject.status,
        priority: editProject.priority,
        targetDate: editProject.targetDate,
        estimatedCost: Number(editProject.estimatedCost || 0),
        notes: editProject.notes,
        categoryId: editProject.categoryId || undefined,
      }),
    });
    if (!resp.ok) {
      showNotice('danger', 'Failed to update project.');
      return;
    }
    const updated = await resp.json();
    setProjects((prev) => prev.map((p) => (p.id === editProjectId ? updated : p)));
    setEditProjectId(null);
    setEditProject(null);
    showNotice('success', 'Project updated.');
  };

  const handleDeleteProject = async (id: number) => {
    const target = projects.find((p) => p.id === id);
    if (!window.confirm(`Delete upgrade project "${target?.title || 'Untitled'}"?`)) return;
    const resp = await fetch(`${SERVER_URL}/upgrades/delete/${id}`, { method: 'DELETE' });
    if (!resp.ok) {
      showNotice('danger', 'Failed to delete project.');
      return;
    }
    setProjects((prev) => prev.filter((p) => p.id !== id));
    showNotice('success', 'Project deleted.');
  };

  const categoryName = (id?: number | null) =>
    categories.find((c) => c.id === id)?.name || 'Uncategorized';

  return (
    <Container style={{ marginTop: '16px' }}>
      <MyNavbar />
      <h3>Upgrade Planning</h3>

      <ToastContainer position="top-end" className="p-3">
        <Toast bg={notice?.variant} onClose={() => setNotice(null)} show={!!notice} delay={2500} autohide>
          <Toast.Body style={{ color: '#fff' }}>{notice?.message}</Toast.Body>
        </Toast>
      </ToastContainer>

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
                          <div style={{ display: 'flex', gap: 6 }}>
                            <Button variant="outline-secondary" size="sm" onClick={() => handleStartEdit(p)}>Edit</Button>
                            <Button variant="outline-danger" size="sm" onClick={() => handleDeleteProject(p.id)}>Delete</Button>
                          </div>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </Table>

              {editProjectId && editProject && (
                <Card style={{ marginTop: 12 }}>
                  <Card.Body>
                    <h6>Edit Project</h6>
                    <Row className="g-2">
                      <Col md={4}>
                        <Form.Control placeholder="Project title" value={editProject.title} onChange={(e) => setEditProject({ ...editProject, title: e.target.value })} />
                      </Col>
                      <Col md={3}>
                        <Form.Select value={editProject.status} onChange={(e) => setEditProject({ ...editProject, status: e.target.value })}>
                          <option value="planned">Planned</option>
                          <option value="in-progress">In Progress</option>
                          <option value="completed">Completed</option>
                          <option value="on-hold">On Hold</option>
                        </Form.Select>
                      </Col>
                      <Col md={3}>
                        <Form.Select value={editProject.priority} onChange={(e) => setEditProject({ ...editProject, priority: e.target.value })}>
                          <option value="low">Low</option>
                          <option value="medium">Medium</option>
                          <option value="high">High</option>
                        </Form.Select>
                      </Col>
                      <Col md={2}>
                        <Form.Control type="date" value={editProject.targetDate || ''} onChange={(e) => setEditProject({ ...editProject, targetDate: e.target.value })} />
                      </Col>
                      <Col md={2}>
                        <Form.Control type="number" step="0.01" placeholder="Estimate" value={editProject.estimatedCost} onChange={(e) => setEditProject({ ...editProject, estimatedCost: Number(e.target.value) })} />
                      </Col>
                      <Col md={4}>
                        <Form.Select value={editProject.categoryId || ''} onChange={(e) => setEditProject({ ...editProject, categoryId: e.target.value ? Number(e.target.value) : undefined })}>
                          <option value="">Category</option>
                          {categories.map((c) => (
                            <option key={c.id} value={c.id}>{c.name}</option>
                          ))}
                        </Form.Select>
                      </Col>
                      <Col md={6}>
                        <Form.Control placeholder="Description" value={editProject.description || ''} onChange={(e) => setEditProject({ ...editProject, description: e.target.value })} />
                      </Col>
                      <Col md={6}>
                        <Form.Control placeholder="Notes" value={editProject.notes || ''} onChange={(e) => setEditProject({ ...editProject, notes: e.target.value })} />
                      </Col>
                      <Col md={12} style={{ display: 'flex', gap: 8 }}>
                        <Button variant="primary" onClick={handleSaveEdit}>Save</Button>
                        <Button variant="outline-secondary" onClick={handleCancelEdit}>Cancel</Button>
                      </Col>
                    </Row>
                  </Card.Body>
                </Card>
              )}

              <Form className="mt-3">
                <Row className="g-2">
                  {projectError && (
                    <Col md={12} style={{ color: '#b02a37' }}>
                      {projectError}
                    </Col>
                  )}
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
