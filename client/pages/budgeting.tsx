import React, { useEffect, useState } from 'react';
import { Button, Card, Col, Container, Form, Row, Table, Toast, ToastContainer } from 'react-bootstrap';
import MyNavbar from '../components/Navbar';
import { SERVER_URL } from '@/pages/_app';

type BudgetScenario = {
  id: number;
  name: string;
  startDate: string;
  horizonMonths: number;
  inflationRate: number;
  isActive: boolean;
  notes: string;
};

type BudgetCategory = {
  id: number;
  name: string;
  assetGroup: string;
  description: string;
  color: string;
};

type PlannedCost = {
  id: number;
  scenarioId?: number | null;
  categoryId?: number | null;
  sourceType: string;
  sourceId?: number | null;
  costDate: string;
  amount: number;
  notes: string;
};

type BudgetSummary = {
  scenario?: BudgetScenario | null;
  horizonMonths: number;
  totalPlanned: number;
  monthlySavings: number;
  upcoming30Days: number;
  upcoming90Days: number;
  plannedCostCount: number;
  categoryTotals: Record<string, number>;
  monthlyBuckets: Array<{ month: string; total: number }>;
};

const BudgetingPage: React.FC = () => {
  const [scenarios, setScenarios] = useState<BudgetScenario[]>([]);
  const [categories, setCategories] = useState<BudgetCategory[]>([]);
  const [plannedCosts, setPlannedCosts] = useState<PlannedCost[]>([]);
  const [selectedScenarioId, setSelectedScenarioId] = useState<number | null>(null);
  const [summary, setSummary] = useState<BudgetSummary | null>(null);
  const [editScenarioId, setEditScenarioId] = useState<number | null>(null);
  const [editScenario, setEditScenario] = useState<BudgetScenario | null>(null);
  const [editCategoryId, setEditCategoryId] = useState<number | null>(null);
  const [editCategory, setEditCategory] = useState<BudgetCategory | null>(null);
  const [editCostId, setEditCostId] = useState<number | null>(null);
  const [editCost, setEditCost] = useState<PlannedCost | null>(null);

  const [newScenario, setNewScenario] = useState({
    name: '',
    startDate: '',
    horizonMonths: 12,
    inflationRate: 0,
    isActive: false,
    notes: '',
  });
  const [scenarioError, setScenarioError] = useState<string>('');

  const [newCategory, setNewCategory] = useState({
    name: '',
    assetGroup: '',
    description: '',
    color: '',
  });
  const [categoryError, setCategoryError] = useState<string>('');

  const [newCost, setNewCost] = useState({
    scenarioId: '',
    categoryId: '',
    sourceType: 'upgrade',
    costDate: '',
    amount: '',
    notes: '',
  });
  const [costError, setCostError] = useState<string>('');
  const [notice, setNotice] = useState<{ variant: 'success' | 'danger'; message: string } | null>(null);
  const showNotice = (variant: 'success' | 'danger', message: string) => setNotice({ variant, message });

  const loadScenarios = async () => {
    try {
      const resp = await fetch(`${SERVER_URL}/budget/scenarios`);
      if (!resp.ok) return;
      const data = await resp.json();
      setScenarios(data);
      if (data.length > 0 && selectedScenarioId === null) {
        const active = data.find((s: BudgetScenario) => s.isActive) || data[0];
        setSelectedScenarioId(active.id);
      }
    } catch (e) {
      console.error('Error loading scenarios', e);
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

  const loadPlannedCosts = async (scenarioId: number | null) => {
    try {
      let url = `${SERVER_URL}/planned-costs`;
      if (scenarioId) {
        url = `${url}?scenarioId=${scenarioId}`;
      }
      const resp = await fetch(url);
      if (!resp.ok) return;
      const data = await resp.json();
      setPlannedCosts(data);
    } catch (e) {
      console.error('Error loading planned costs', e);
    }
  };

  useEffect(() => {
    loadScenarios();
    loadCategories();
  }, []);

  useEffect(() => {
    loadPlannedCosts(selectedScenarioId);
  }, [selectedScenarioId]);

  useEffect(() => {
    const loadSummary = async () => {
      try {
        let url = `${SERVER_URL}/budget/summary`;
        if (selectedScenarioId) {
          url = `${url}?scenarioId=${selectedScenarioId}`;
        }
        const resp = await fetch(url);
        if (!resp.ok) return;
        const data = await resp.json();
        setSummary(data);
      } catch (e) {
        console.error('Error loading budget summary', e);
      }
    };

    loadSummary();
  }, [selectedScenarioId]);

  const handleAddScenario = async () => {
    if (!newScenario.name) {
      setScenarioError('Scenario name is required.');
      return;
    }
    if (!newScenario.horizonMonths || newScenario.horizonMonths <= 0) {
      setNewScenario({ ...newScenario, horizonMonths: 12 });
    }
    setScenarioError('');
    const resp = await fetch(`${SERVER_URL}/budget/scenarios/add`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: newScenario.name,
        startDate: newScenario.startDate,
        horizonMonths: Number(newScenario.horizonMonths || 12),
        inflationRate: Number(newScenario.inflationRate),
        isActive: !!newScenario.isActive,
        notes: newScenario.notes,
      }),
    });
    if (!resp.ok) {
      showNotice('danger', 'Failed to add scenario.');
      return;
    }
    const created = await resp.json();
    setScenarios((prev) => [...prev, created]);
    if (selectedScenarioId === null) setSelectedScenarioId(created.id);
    setNewScenario({ name: '', startDate: '', horizonMonths: 12, inflationRate: 0, isActive: false, notes: '' });
    showNotice('success', 'Scenario added.');
  };

  const handleStartScenarioEdit = (scenario: BudgetScenario) => {
    setEditScenarioId(scenario.id);
    setEditScenario({ ...scenario });
  };

  const handleCancelScenarioEdit = () => {
    setEditScenarioId(null);
    setEditScenario(null);
  };

  const handleSaveScenario = async () => {
    if (!editScenario || !editScenarioId) return;
    if (!editScenario.name) {
      setScenarioError('Scenario name is required.');
      return;
    }
    const resp = await fetch(`${SERVER_URL}/budget/scenarios/update/${editScenarioId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: editScenario.name,
        startDate: editScenario.startDate,
        horizonMonths: Number(editScenario.horizonMonths || 12),
        inflationRate: Number(editScenario.inflationRate || 0),
        isActive: !!editScenario.isActive,
        notes: editScenario.notes,
      }),
    });
    if (!resp.ok) {
      showNotice('danger', 'Failed to update scenario.');
      return;
    }
    const updated = await resp.json();
    setScenarios((prev) => prev.map((s) => (s.id === editScenarioId ? updated : s)));
    setEditScenarioId(null);
    setEditScenario(null);
    showNotice('success', 'Scenario updated.');
  };

  const handleAddCategory = async () => {
    if (!newCategory.name) {
      setCategoryError('Category name is required.');
      return;
    }
    setCategoryError('');
    const resp = await fetch(`${SERVER_URL}/budget/categories/add`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(newCategory),
    });
    if (!resp.ok) {
      showNotice('danger', 'Failed to add category.');
      return;
    }
    const created = await resp.json();
    setCategories((prev) => [...prev, created]);
    setNewCategory({ name: '', assetGroup: '', description: '', color: '' });
    showNotice('success', 'Category added.');
  };

  const handleStartCategoryEdit = (category: BudgetCategory) => {
    setEditCategoryId(category.id);
    setEditCategory({ ...category });
  };

  const handleCancelCategoryEdit = () => {
    setEditCategoryId(null);
    setEditCategory(null);
  };

  const handleSaveCategory = async () => {
    if (!editCategory || !editCategoryId) return;
    if (!editCategory.name) {
      setCategoryError('Category name is required.');
      return;
    }
    const resp = await fetch(`${SERVER_URL}/budget/categories/update/${editCategoryId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: editCategory.name,
        assetGroup: editCategory.assetGroup,
        description: editCategory.description,
        color: editCategory.color,
      }),
    });
    if (!resp.ok) {
      showNotice('danger', 'Failed to update category.');
      return;
    }
    const updated = await resp.json();
    setCategories((prev) => prev.map((c) => (c.id === editCategoryId ? updated : c)));
    setEditCategoryId(null);
    setEditCategory(null);
    showNotice('success', 'Category updated.');
  };

  const handleAddCost = async () => {
    if (!newCost.costDate || !newCost.amount) {
      setCostError('Cost date and amount are required.');
      return;
    }
    if (Number(newCost.amount) <= 0) {
      setCostError('Amount must be greater than zero.');
      return;
    }
    setCostError('');
    const resp = await fetch(`${SERVER_URL}/planned-costs/add`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        scenarioId: newCost.scenarioId ? Number(newCost.scenarioId) : undefined,
        categoryId: newCost.categoryId ? Number(newCost.categoryId) : undefined,
        sourceType: newCost.sourceType,
        costDate: newCost.costDate,
        amount: Number(newCost.amount),
        notes: newCost.notes,
      }),
    });
    if (!resp.ok) {
      showNotice('danger', 'Failed to add planned cost.');
      return;
    }
    const created = await resp.json();
    setPlannedCosts((prev) => [...prev, created]);
    setNewCost({ scenarioId: '', categoryId: '', sourceType: 'upgrade', costDate: '', amount: '', notes: '' });
    showNotice('success', 'Planned cost added.');
  };

  const handleStartCostEdit = (cost: PlannedCost) => {
    setEditCostId(cost.id);
    setEditCost({ ...cost });
  };

  const handleCancelCostEdit = () => {
    setEditCostId(null);
    setEditCost(null);
  };

  const handleSaveCost = async () => {
    if (!editCost || !editCostId) return;
    if (!editCost.costDate || !editCost.amount || Number(editCost.amount) <= 0) {
      setCostError('Cost date and positive amount are required.');
      return;
    }
    setCostError('');
    const resp = await fetch(`${SERVER_URL}/planned-costs/update/${editCostId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        scenarioId: editCost.scenarioId || undefined,
        categoryId: editCost.categoryId || undefined,
        sourceType: editCost.sourceType,
        sourceId: editCost.sourceId || undefined,
        costDate: editCost.costDate,
        amount: Number(editCost.amount || 0),
        notes: editCost.notes,
      }),
    });
    if (!resp.ok) {
      showNotice('danger', 'Failed to update planned cost.');
      return;
    }
    const updated = await resp.json();
    setPlannedCosts((prev) => prev.map((c) => (c.id === editCostId ? updated : c)));
    setEditCostId(null);
    setEditCost(null);
    showNotice('success', 'Planned cost updated.');
  };

  const handleDeleteCost = async (id: number) => {
    const target = plannedCosts.find((c) => c.id === id);
    if (!window.confirm(`Delete planned cost on ${target?.costDate || 'unknown date'}?`)) return;
    const resp = await fetch(`${SERVER_URL}/planned-costs/delete/${id}`, { method: 'DELETE' });
    if (!resp.ok) {
      showNotice('danger', 'Failed to delete planned cost.');
      return;
    }
    setPlannedCosts((prev) => prev.filter((c) => c.id !== id));
    showNotice('success', 'Planned cost deleted.');
  };

  const handleDeleteScenario = async (id: number) => {
    const target = scenarios.find((s) => s.id === id);
    if (!window.confirm(`Delete scenario "${target?.name || 'Untitled'}"?`)) return;
    const resp = await fetch(`${SERVER_URL}/budget/scenarios/delete/${id}`, { method: 'DELETE' });
    if (!resp.ok) {
      showNotice('danger', 'Failed to delete scenario.');
      return;
    }
    setScenarios((prev) => prev.filter((s) => s.id !== id));
    if (selectedScenarioId === id) setSelectedScenarioId(null);
    showNotice('success', 'Scenario deleted.');
  };

  const handleDeleteCategory = async (id: number) => {
    const target = categories.find((c) => c.id === id);
    if (!window.confirm(`Delete category "${target?.name || 'Untitled'}"?`)) return;
    const resp = await fetch(`${SERVER_URL}/budget/categories/delete/${id}`, { method: 'DELETE' });
    if (!resp.ok) {
      showNotice('danger', 'Failed to delete category.');
      return;
    }
    setCategories((prev) => prev.filter((c) => c.id !== id));
    showNotice('success', 'Category deleted.');
  };

  const categoryName = (id?: number | null) =>
    categories.find((c) => c.id === id)?.name || 'Uncategorized';

  return (
    <Container style={{ marginTop: '16px' }}>
      <MyNavbar />
      <h3>Budgeting</h3>

      <ToastContainer position="top-end" className="p-3">
        <Toast bg={notice?.variant} onClose={() => setNotice(null)} show={!!notice} delay={2500} autohide>
          <Toast.Body style={{ color: '#fff' }}>{notice?.message}</Toast.Body>
        </Toast>
      </ToastContainer>

      {!selectedScenarioId && (
        <Card style={{ marginBottom: '12px' }}>
          <Card.Body>
            Create a scenario to enable budgeting summaries and horizon-based views.
          </Card.Body>
        </Card>
      )}

      <Row className="g-3" style={{ marginBottom: '12px' }}>
        <Col md={4}>
          <Card>
            <Card.Body>
              <div style={{ fontSize: '0.9rem', color: '#6c757d' }}>Monthly Savings Target</div>
              <div style={{ fontSize: '1.3rem', fontWeight: 600 }}>${(summary?.monthlySavings || 0).toFixed(2)}/mo</div>
              <div style={{ fontSize: '0.85rem' }}>{summary?.plannedCostCount || 0} items in plan</div>
            </Card.Body>
          </Card>
        </Col>
        <Col md={4}>
          <Card>
            <Card.Body>
              <div style={{ fontSize: '0.9rem', color: '#6c757d' }}>Upcoming (30 days)</div>
              <div style={{ fontSize: '1.3rem', fontWeight: 600 }}>${(summary?.upcoming30Days || 0).toFixed(2)}</div>
              <div style={{ fontSize: '0.85rem' }}>Next 90 days: ${(summary?.upcoming90Days || 0).toFixed(2)}</div>
            </Card.Body>
          </Card>
        </Col>
        <Col md={4}>
          <Card>
            <Card.Body>
              <div style={{ fontSize: '0.9rem', color: '#6c757d' }}>Total Planned</div>
              <div style={{ fontSize: '1.3rem', fontWeight: 600 }}>${(summary?.totalPlanned || 0).toFixed(2)}</div>
              <div style={{ fontSize: '0.85rem' }}>{summary?.horizonMonths || 0} month horizon</div>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      <Row className="g-3">
        <Col lg={6}>
          <Card>
            <Card.Body>
              <h5>Scenarios</h5>
              <Table striped bordered hover size="sm">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Start</th>
                    <th>Horizon</th>
                    <th>Active</th>
                  </tr>
                </thead>
                <tbody>
                  {scenarios.length === 0 ? (
                    <tr>
                      <td colSpan={4} style={{ textAlign: 'center' }}>No scenarios</td>
                    </tr>
                  ) : (
                    scenarios.map((s) => (
                      <tr key={s.id} onClick={() => setSelectedScenarioId(s.id)} style={{ cursor: 'pointer' }}>
                        <td>{s.name}</td>
                        <td>{s.startDate || '-'}</td>
                        <td>{s.horizonMonths} mo</td>
                        <td style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                          <span>{s.isActive ? 'Yes' : 'No'}</span>
                          <Button variant="outline-secondary" size="sm" onClick={(e) => { e.stopPropagation(); handleStartScenarioEdit(s); }}>Edit</Button>
                          <Button variant="outline-danger" size="sm" onClick={(e) => { e.stopPropagation(); handleDeleteScenario(s.id); }}>Delete</Button>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </Table>

              {editScenarioId && editScenario && (
                <Card style={{ marginTop: 12 }}>
                  <Card.Body>
                    <h6>Edit Scenario</h6>
                    <Row className="g-2">
                      <Col md={6}>
                        <Form.Control placeholder="Scenario name" value={editScenario.name} onChange={(e) => setEditScenario({ ...editScenario, name: e.target.value })} />
                      </Col>
                      <Col md={6}>
                        <Form.Control type="date" value={editScenario.startDate || ''} onChange={(e) => setEditScenario({ ...editScenario, startDate: e.target.value })} />
                      </Col>
                      <Col md={4}>
                        <Form.Control type="number" min={1} placeholder="Horizon (months)" value={editScenario.horizonMonths} onChange={(e) => setEditScenario({ ...editScenario, horizonMonths: Number(e.target.value) })} />
                      </Col>
                      <Col md={4}>
                        <Form.Control type="number" step="0.1" placeholder="Inflation %" value={editScenario.inflationRate} onChange={(e) => setEditScenario({ ...editScenario, inflationRate: Number(e.target.value) })} />
                      </Col>
                      <Col md={4}>
                        <Form.Check type="checkbox" label="Active" checked={!!editScenario.isActive} onChange={(e) => setEditScenario({ ...editScenario, isActive: e.target.checked })} />
                      </Col>
                      <Col md={12}>
                        <Form.Control placeholder="Notes" value={editScenario.notes || ''} onChange={(e) => setEditScenario({ ...editScenario, notes: e.target.value })} />
                      </Col>
                      <Col md={12} style={{ display: 'flex', gap: 8 }}>
                        <Button variant="primary" onClick={handleSaveScenario}>Save</Button>
                        <Button variant="outline-secondary" onClick={handleCancelScenarioEdit}>Cancel</Button>
                      </Col>
                    </Row>
                  </Card.Body>
                </Card>
              )}

              <Form className="mt-3">
                <Row className="g-2">
                  {scenarioError && (
                    <Col md={12} style={{ color: '#b02a37' }}>
                      {scenarioError}
                    </Col>
                  )}
                  <Col md={6}>
                    <Form.Control placeholder="Scenario name" value={newScenario.name} onChange={(e) => setNewScenario({ ...newScenario, name: e.target.value })} />
                  </Col>
                  <Col md={6}>
                    <Form.Control type="date" value={newScenario.startDate} onChange={(e) => setNewScenario({ ...newScenario, startDate: e.target.value })} />
                  </Col>
                  <Col md={4}>
                    <Form.Control type="number" min={1} placeholder="Horizon (months)" value={newScenario.horizonMonths} onChange={(e) => setNewScenario({ ...newScenario, horizonMonths: Number(e.target.value) })} />
                  </Col>
                  <Col md={4}>
                    <Form.Control type="number" step="0.1" placeholder="Inflation %" value={newScenario.inflationRate} onChange={(e) => setNewScenario({ ...newScenario, inflationRate: Number(e.target.value) })} />
                  </Col>
                  <Col md={4}>
                    <Form.Check type="checkbox" label="Active" checked={newScenario.isActive} onChange={(e) => setNewScenario({ ...newScenario, isActive: e.target.checked })} />
                  </Col>
                  <Col md={12}>
                    <Form.Control placeholder="Notes" value={newScenario.notes} onChange={(e) => setNewScenario({ ...newScenario, notes: e.target.value })} />
                  </Col>
                  <Col md={12}>
                    <Button onClick={handleAddScenario}>Add Scenario</Button>
                  </Col>
                </Row>
              </Form>
            </Card.Body>
          </Card>
        </Col>

        <Col lg={6}>
          <Card>
            <Card.Body>
              <h5>Categories</h5>
              <Table striped bordered hover size="sm">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Asset Group</th>
                    <th>Color</th>
                  </tr>
                </thead>
                <tbody>
                  {categories.length === 0 ? (
                    <tr>
                      <td colSpan={3} style={{ textAlign: 'center' }}>No categories</td>
                    </tr>
                  ) : (
                    categories.map((c) => (
                      <tr key={c.id}>
                        <td>{c.name}</td>
                        <td>{c.assetGroup || '-'}</td>
                        <td style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                          <span>{c.color || '-'}</span>
                          <Button variant="outline-secondary" size="sm" onClick={() => handleStartCategoryEdit(c)}>Edit</Button>
                          <Button variant="outline-danger" size="sm" onClick={() => handleDeleteCategory(c.id)}>Delete</Button>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </Table>

              {editCategoryId && editCategory && (
                <Card style={{ marginTop: 12 }}>
                  <Card.Body>
                    <h6>Edit Category</h6>
                    <Row className="g-2">
                      <Col md={6}>
                        <Form.Control placeholder="Category name" value={editCategory.name} onChange={(e) => setEditCategory({ ...editCategory, name: e.target.value })} />
                      </Col>
                      <Col md={6}>
                        <Form.Control placeholder="Asset group" value={editCategory.assetGroup || ''} onChange={(e) => setEditCategory({ ...editCategory, assetGroup: e.target.value })} />
                      </Col>
                      <Col md={6}>
                        <Form.Control placeholder="Color" value={editCategory.color || ''} onChange={(e) => setEditCategory({ ...editCategory, color: e.target.value })} />
                      </Col>
                      <Col md={6}>
                        <Form.Control placeholder="Description" value={editCategory.description || ''} onChange={(e) => setEditCategory({ ...editCategory, description: e.target.value })} />
                      </Col>
                      <Col md={12} style={{ display: 'flex', gap: 8 }}>
                        <Button variant="primary" onClick={handleSaveCategory}>Save</Button>
                        <Button variant="outline-secondary" onClick={handleCancelCategoryEdit}>Cancel</Button>
                      </Col>
                    </Row>
                  </Card.Body>
                </Card>
              )}

              <Form className="mt-3">
                <Row className="g-2">
                  {categoryError && (
                    <Col md={12} style={{ color: '#b02a37' }}>
                      {categoryError}
                    </Col>
                  )}
                  <Col md={6}>
                    <Form.Control placeholder="Category name" value={newCategory.name} onChange={(e) => setNewCategory({ ...newCategory, name: e.target.value })} />
                  </Col>
                  <Col md={6}>
                    <Form.Control placeholder="Asset group" value={newCategory.assetGroup} onChange={(e) => setNewCategory({ ...newCategory, assetGroup: e.target.value })} />
                  </Col>
                  <Col md={6}>
                    <Form.Control placeholder="Color" value={newCategory.color} onChange={(e) => setNewCategory({ ...newCategory, color: e.target.value })} />
                  </Col>
                  <Col md={6}>
                    <Form.Control placeholder="Description" value={newCategory.description} onChange={(e) => setNewCategory({ ...newCategory, description: e.target.value })} />
                  </Col>
                  <Col md={12}>
                    <Button onClick={handleAddCategory}>Add Category</Button>
                  </Col>
                </Row>
              </Form>
            </Card.Body>
          </Card>
        </Col>

        <Col lg={6}>
          <Card>
            <Card.Body>
              <h5>Category Breakdown</h5>
              <Table striped bordered hover size="sm">
                <thead>
                  <tr>
                    <th>Category</th>
                    <th>Total</th>
                  </tr>
                </thead>
                <tbody>
                  {!summary || Object.keys(summary.categoryTotals || {}).length === 0 ? (
                    <tr>
                      <td colSpan={2} style={{ textAlign: 'center' }}>No category data</td>
                    </tr>
                  ) : (
                    Object.entries(summary.categoryTotals)
                      .sort((a, b) => b[1] - a[1])
                      .map(([name, total]) => (
                        <tr key={name}>
                          <td>{name}</td>
                          <td>${Number(total || 0).toFixed(2)}</td>
                        </tr>
                      ))
                  )}
                </tbody>
              </Table>
            </Card.Body>
          </Card>
        </Col>

        <Col lg={6}>
          <Card>
            <Card.Body>
              <h5>Monthly Buckets</h5>
              <Table striped bordered hover size="sm">
                <thead>
                  <tr>
                    <th>Month</th>
                    <th>Total</th>
                  </tr>
                </thead>
                <tbody>
                  {!summary || (summary.monthlyBuckets || []).length === 0 ? (
                    <tr>
                      <td colSpan={2} style={{ textAlign: 'center' }}>No monthly data</td>
                    </tr>
                  ) : (
                    summary.monthlyBuckets.map((m) => (
                      <tr key={m.month}>
                        <td>{m.month}</td>
                        <td>${Number(m.total || 0).toFixed(2)}</td>
                      </tr>
                    ))
                  )}
                </tbody>
              </Table>
            </Card.Body>
          </Card>
        </Col>

        <Col lg={12}>
          <Card>
            <Card.Body>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <h5>Planned Costs</h5>
                <Form.Select style={{ maxWidth: 260 }} value={selectedScenarioId ?? ''} onChange={(e) => setSelectedScenarioId(e.target.value ? Number(e.target.value) : null)}>
                  <option value="">All scenarios</option>
                  {scenarios.map((s) => (
                    <option key={s.id} value={s.id}>{s.name}</option>
                  ))}
                </Form.Select>
              </div>
              <Table striped bordered hover size="sm" className="mt-2">
                <thead>
                  <tr>
                    <th>Date</th>
                    <th>Amount</th>
                    <th>Category</th>
                    <th>Type</th>
                    <th>Notes</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  {plannedCosts.length === 0 ? (
                    <tr>
                      <td colSpan={6} style={{ textAlign: 'center' }}>No planned costs</td>
                    </tr>
                  ) : (
                    plannedCosts.map((c) => (
                      <tr key={c.id}>
                        <td>{c.costDate || '-'}</td>
                        <td>${c.amount?.toFixed(2)}</td>
                        <td>{categoryName(c.categoryId)}</td>
                        <td>{c.sourceType || '-'}</td>
                        <td>{c.notes || '-'}</td>
                        <td style={{ width: 70 }}>
                          <div style={{ display: 'flex', gap: 6 }}>
                            <Button variant="outline-secondary" size="sm" onClick={() => handleStartCostEdit(c)}>Edit</Button>
                            <Button variant="outline-danger" size="sm" onClick={() => handleDeleteCost(c.id)}>Delete</Button>
                          </div>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </Table>

              {editCostId && editCost && (
                <Card style={{ marginTop: 12 }}>
                  <Card.Body>
                    <h6>Edit Planned Cost</h6>
                    <Row className="g-2">
                      <Col md={3}>
                        <Form.Select value={editCost.scenarioId || ''} onChange={(e) => setEditCost({ ...editCost, scenarioId: e.target.value ? Number(e.target.value) : null })}>
                          <option value="">Scenario</option>
                          {scenarios.map((s) => (
                            <option key={s.id} value={s.id}>{s.name}</option>
                          ))}
                        </Form.Select>
                      </Col>
                      <Col md={3}>
                        <Form.Select value={editCost.categoryId || ''} onChange={(e) => setEditCost({ ...editCost, categoryId: e.target.value ? Number(e.target.value) : null })}>
                          <option value="">Category</option>
                          {categories.map((c) => (
                            <option key={c.id} value={c.id}>{c.name}</option>
                          ))}
                        </Form.Select>
                      </Col>
                      <Col md={2}>
                        <Form.Select value={editCost.sourceType || 'upgrade'} onChange={(e) => setEditCost({ ...editCost, sourceType: e.target.value })}>
                          <option value="upgrade">Upgrade</option>
                          <option value="repair">Repair</option>
                          <option value="maintenance">Maintenance</option>
                          <option value="recurring">Recurring</option>
                          <option value="other">Other</option>
                        </Form.Select>
                      </Col>
                      <Col md={2}>
                        <Form.Control type="date" value={editCost.costDate || ''} onChange={(e) => setEditCost({ ...editCost, costDate: e.target.value })} />
                      </Col>
                      <Col md={2}>
                        <Form.Control type="number" step="0.01" placeholder="Amount" value={editCost.amount} onChange={(e) => setEditCost({ ...editCost, amount: Number(e.target.value) })} />
                      </Col>
                      <Col md={12}>
                        <Form.Control placeholder="Notes" value={editCost.notes || ''} onChange={(e) => setEditCost({ ...editCost, notes: e.target.value })} />
                      </Col>
                      <Col md={12} style={{ display: 'flex', gap: 8 }}>
                        <Button variant="primary" onClick={handleSaveCost}>Save</Button>
                        <Button variant="outline-secondary" onClick={handleCancelCostEdit}>Cancel</Button>
                      </Col>
                    </Row>
                  </Card.Body>
                </Card>
              )}

              <Form className="mt-3">
                <Row className="g-2">
                  {costError && (
                    <Col md={12} style={{ color: '#b02a37' }}>
                      {costError}
                    </Col>
                  )}
                  <Col md={3}>
                    <Form.Select value={newCost.scenarioId} onChange={(e) => setNewCost({ ...newCost, scenarioId: e.target.value })}>
                      <option value="">Scenario</option>
                      {scenarios.map((s) => (
                        <option key={s.id} value={s.id}>{s.name}</option>
                      ))}
                    </Form.Select>
                  </Col>
                  <Col md={3}>
                    <Form.Select value={newCost.categoryId} onChange={(e) => setNewCost({ ...newCost, categoryId: e.target.value })}>
                      <option value="">Category</option>
                      {categories.map((c) => (
                        <option key={c.id} value={c.id}>{c.name}</option>
                      ))}
                    </Form.Select>
                  </Col>
                  <Col md={2}>
                    <Form.Select value={newCost.sourceType} onChange={(e) => setNewCost({ ...newCost, sourceType: e.target.value })}>
                      <option value="upgrade">Upgrade</option>
                      <option value="repair">Repair</option>
                      <option value="maintenance">Maintenance</option>
                      <option value="recurring">Recurring</option>
                      <option value="other">Other</option>
                    </Form.Select>
                  </Col>
                  <Col md={2}>
                    <Form.Control type="date" value={newCost.costDate} onChange={(e) => setNewCost({ ...newCost, costDate: e.target.value })} />
                  </Col>
                  <Col md={2}>
                    <Form.Control type="number" step="0.01" placeholder="Amount" value={newCost.amount} onChange={(e) => setNewCost({ ...newCost, amount: e.target.value })} />
                  </Col>
                  <Col md={10}>
                    <Form.Control placeholder="Notes" value={newCost.notes} onChange={(e) => setNewCost({ ...newCost, notes: e.target.value })} />
                  </Col>
                  <Col md={2}>
                    <Button onClick={handleAddCost}>Add Cost</Button>
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

export default BudgetingPage;
