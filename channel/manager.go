func (m *Manager) SaveConfig() error {
	return utils.SaveConfig("channel", m.Sequence)
}

func (m *Manager) saveAndReload() error {
	if err := m.SaveConfig(); err != nil {
		return err
	}

	m.Load()
	return nil
}

func (m *Manager) CreateChannel(channel *Channel) error {
	channel.Id = m.GetMaxId() + 1
	m.Sequence = append(m.Sequence, channel)
	return m.saveAndReload()
}

func (m *Manager) UpdateChannel(id int, channel *Channel) error {
	for i, item := range m.Sequence {
		if item.Id == id {
			m.Sequence[i] = channel
			return m.saveAndReload()
		}
	}

	return errors.New("channel not found")
}

func (m *Manager) DeleteChannel(id int) error {
	for i, item := range m.Sequence {
		if item.Id == id {
			m.Sequence = append(m.Sequence[:i], m.Sequence[i+1:]...)
			return m.saveAndReload()
		}
	}

	return errors.New("channel not found")
}

func (m *Manager) ActivateChannel(id int) error {
	for i, item := range m.Sequence {
		if item.Id == id {
			m.Sequence[i].State = true
			return m.saveAndReload()
		}
	}

	return errors.New("channel not found")
}

func (m *Manager) DeactivateChannel(id int) error {
	for i, item := range m.Sequence {
		if item.Id == id {
			m.Sequence[i].State = false
			return m.saveAndReload()
		}
	}

	return errors.New("channel not found")
}
